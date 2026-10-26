#!/usr/bin/env python3
"""Manual same-capability rejection QA. Default watch is offline; no dispatch.

Run watch outside the Agent process tree, then send inside its approved run.
Credentials remain in process memory, private local IPC, and the CLI child env.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import socket
import subprocess
import sys
import time

# Agent execution must not create source-tree artifacts while arming the probe.
sys.dont_write_bytecode = True

from ended_run_ipc import (Failure, daemon_status, disable_core_dumps, private_path,
                          read_json, remaining, require, same_user_peer, send_json)


FIELDS = ("RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID", "RIIDO_WORKSPACE_ID")
FORBIDDEN = {"RIIDO_DEVICE_SECRET", "RIIDO_SAAS_TOKEN", "RIIDO_TOKEN", "RIIDO_LEASE_TOKEN", "LEASE_TOKEN"}


def identity(status):
    require(status.get("schema_version") == "riido-agent-daemon-status.v1", "STATUS_SCHEMA")
    values = tuple(status.get(k) for k in ("pid", "started_at", "daemon_version"))
    require(type(values[0]) is int and values[0] > 0 and all(values[1:]), "STATUS_IDENTITY_MISSING")
    return values


def active_execution(status, binding):
    metrics = status.get("metrics", {})
    require(status.get("ready") is True and metrics.get("runtime_count", 0) > 0
            and metrics.get("runtime_count") == metrics.get("runtime_responding"), "STATUS_INCONCLUSIVE")
    runtimes, executions = status.get("runtimes"), status.get("executions")
    require(isinstance(runtimes, list) and isinstance(executions, list), "STATUS_INCONCLUSIVE")
    runtime = [r for r in runtimes if r.get("runtime_id") == binding["runtime_id"]]
    require(len(runtime) == 1 and isinstance(runtime[0].get("running_tasks"), list), "STATUS_INCONCLUSIVE")
    matches = [e for e in executions if e.get("assignment_id") == binding["assignment_id"]]
    for match in matches:
        require(all(match.get(k) == binding[k] for k in ("runtime_id", "task_id", "workspace_id")), "EXECUTION_SCOPE_CHANGED")
    native = any(t.get("task_id") == binding["assignment_id"] for t in runtime[0]["running_tasks"])
    return bool(matches) or native


def observe_ended_run(cli, get_status, task, workspace, timeout, clock=time.monotonic, sleep=time.sleep, baseline_ready=lambda: None):
    code, initial = cli(["whoami", "--json"])
    require(code == 0 and initial.get("error") is None, "ACTIVE_CAPABILITY_NOT_CONFIRMED")
    data = initial.get("data") or {}
    binding = data.get("run") or {}
    require(data.get("mode") == "agent" and binding.get("task_id") == task
            and binding.get("workspace_id") == workspace
            and all(isinstance(binding.get(k), str) and binding[k]
                    for k in ("assignment_id", "runtime_id")), "ACTIVE_CAPABILITY_SCOPE")
    first = get_status()
    daemon = identity(first)
    require(active_execution(first, binding) and any(e.get("assignment_id") == binding["assignment_id"]
            for e in first["executions"]), "ACTIVE_EXECUTION_NOT_OBSERVED")
    baseline_ready()
    deadline, absent, observations = clock() + timeout, 0, 1
    while clock() < deadline:
        sleep(1)
        current = get_status()
        observations += 1
        require(identity(current) == daemon, "DAEMON_CHANGED")
        absent = 0 if active_execution(current, binding) else absent + 1
        if absent < 2:
            continue
        # whoami is answered entirely by the broker, so a server-side denial
        # cannot masquerade as revoked local authority. Also check the read leaf.
        for command in (["whoami", "--json"], ["context", "get", "--json"]):
            code, reply = cli(command)
            require(code == 4 and isinstance(reply.get("error"), dict)
                    and reply["error"].get("code") == "PERMISSION_DENIED", "ENDED_CAPABILITY_NOT_DENIED")
            require(reply.get("data") is None and reply.get("sources") == [], "DENIAL_DATA_LEAK")
        final = get_status()
        require(identity(final) == daemon, "DAEMON_CHANGED")
        require(not active_execution(final, binding), "EXECUTION_REAPPEARED")
        return {"report_schema": "1", "status": "PASS", "staging_acceptance": "NOT_COMPLETE",
                "active_capability_confirmed": True, "execution_ended_observed": True,
                "same_capability_denied": True, "local_whoami_denied": True, "daemon_unchanged": True,
                "cli_exit": 4, "error_code": "PERMISSION_DENIED", "status_observations": observations}
    raise Failure("EXECUTION_END_TIMEOUT")


def cli_runner(binary, payload, deadline=None):
    token = payload["RIIDO_RUN_TOKEN"].encode()
    env = {"PATH": os.defpath, **payload}

    def run(args):
        budget = remaining(deadline, 30) if deadline is not None else 30
        result = subprocess.run([binary, *args], env=env, capture_output=True, timeout=budget, check=False)
        require(token not in result.stdout and token not in result.stderr, "CAPABILITY_OUTPUT_LEAK")
        require(not result.stderr.strip(), "CLI_UNEXPECTED_STDERR")
        require(len(result.stdout) <= 2 * 1024 * 1024, "CLI_RESPONSE_TOO_LARGE")
        try:
            value = json.loads(result.stdout)
        except (ValueError, UnicodeError):
            raise Failure("CLI_INVALID_JSON") from None
        require(isinstance(value, dict) and value.get("schema_version") == "1", "CLI_ENVELOPE_SCHEMA")
        return result.returncode, value
    return run


def safe_environment(env):
    require(not any(k.upper() in FORBIDDEN and value for k, value in env.items()), "PRIVILEGED_CREDENTIAL_PRESENT")


def watch(args):
    deadline = time.monotonic() + args.timeout
    require(args.approval and args.daemon_build, "APPROVAL_REFERENCES_REQUIRED")
    require(args.environment in {"isolated", "staging"}, "ENVIRONMENT_REQUIRED")
    require(args.expected_task and args.daemon_socket and args.cli, "FIXTURE_ARGUMENTS_REQUIRED")
    require(5 <= args.timeout <= 600, "TIMEOUT_BOUNDS")
    require(not os.environ.get("RIIDO_RUN_TOKEN"), "WATCH_MUST_BE_OUTSIDE_AGENT")
    safe_environment(os.environ)
    require(re.fullmatch(r"[0-9a-f]{64}", args.cli_sha256 or ""), "CLI_HASH_REQUIRED")
    require(Path(args.cli).is_absolute() and hashlib.sha256(Path(args.cli).read_bytes()).hexdigest() == args.cli_sha256,
            "CLI_HASH_MISMATCH")
    path = private_path(args.handoff)
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as listener:
        listener.bind(path)
        try:
            os.chmod(path, 0o600)
            listener.listen(1)
            listener.settimeout(remaining(deadline, args.timeout))
            print(json.dumps({"event": "READY", "mode": args.environment}), flush=True)
            conn, _ = listener.accept()
            with conn:
                conn.settimeout(remaining(deadline, 40))
                same_user_peer(conn)
                payload = read_json(conn, deadline=deadline)
                require(set(payload) == set(FIELDS) and all(isinstance(v, str) and v for v in payload.values()), "HANDOFF_FIELDS")
                require(re.fullmatch(r"rrc_[A-Za-z0-9_-]{20,100}", payload["RIIDO_RUN_TOKEN"]), "CAPABILITY_FORMAT")
                require(payload["RIIDO_DAEMON_SOCKET"] == args.daemon_socket
                        and payload["RIIDO_TASK_ID"] == args.expected_task, "HANDOFF_SCOPE_MISMATCH")
                report = observe_ended_run(cli_runner(args.cli, payload, deadline), lambda: daemon_status(args.daemon_socket, deadline),
                                          args.expected_task, payload["RIIDO_WORKSPACE_ID"], remaining(deadline, args.timeout),
                                          baseline_ready=lambda: send_json(conn, {"accepted": True}))
                report["mode"] = args.environment
                return report
        finally:
            Path(path).unlink(missing_ok=True)


def send(args):
    deadline = time.monotonic() + 40
    safe_environment(os.environ)
    payload = {key: os.environ.get(key, "") for key in FIELDS}
    require(all(payload.values()), "AGENT_ENV_REQUIRED")
    path = private_path(args.handoff, existing=True)
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
        conn.settimeout(remaining(deadline, 40))
        conn.connect(path)
        same_user_peer(conn)
        send_json(conn, payload)
        require(read_json(conn, deadline=deadline) == {"accepted": True}, "BASELINE_NOT_ACKNOWLEDGED")
    return {"report_schema": "1", "status": "ARMED", "staging_acceptance": "NOT_COMPLETE"}


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("watch", "send"))
    parser.add_argument("--handoff", help="Private Unix socket in an owned, mode-0700 directory")
    parser.add_argument("--execute", action="store_true")
    parser.add_argument("--approval", help="Non-secret approval reference")
    parser.add_argument("--environment", choices=("isolated", "staging"))
    parser.add_argument("--daemon-build", help="Non-secret approved build reference")
    parser.add_argument("--daemon-socket")
    parser.add_argument("--cli", help="Absolute path to the approved riido binary")
    parser.add_argument("--cli-sha256")
    parser.add_argument("--expected-task")
    parser.add_argument("--timeout", type=int, default=180)
    args = parser.parse_args(argv)
    if args.mode == "watch" and not args.execute:
        print(json.dumps({"report_schema": "1", "mode": "offline_plan", "status": "NOT_RUN", "staging_acceptance": "NOT_COMPLETE"}))
        return 0
    try:
        disable_core_dumps()
        require(args.handoff, "HANDOFF_REQUIRED")
        report = watch(args) if args.mode == "watch" else send(args)
        code = 0
    except Failure as exc:
        report, code = {"status": "FAIL", "reason": str(exc), "staging_acceptance": "NOT_COMPLETE"}, 1
    except (OSError, ValueError, TypeError, KeyError, subprocess.SubprocessError):
        report, code = {"status": "FAIL", "reason": "LOCAL_PROBE_ERROR", "staging_acceptance": "NOT_COMPLETE"}, 1
    print(json.dumps(report), flush=True)
    return code


if __name__ == "__main__":
    raise SystemExit(main())
