#!/usr/bin/env python3
"""Bounded, read-only CLI QA inside an already approved staging Agent run.

Default invocation is offline. No dispatch, installation, server auth, or retry.
Only fixed check names, counts, and booleans enter the JSON report.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys


MANUAL_CHECKS = [
    "staging_endpoint_and_running_daemon_build",
    "prompt_and_cli_read_records_and_sse",
    "sse_visibility_and_payload_exclusions",
    "ended_run_capability_rejected",
    "daemon_and_provider_logs_have_no_credentials",
    "snapshot_staleness_and_next_assignment_refresh",
    "runtime_installation_and_socket_matrix",
]
FORBIDDEN_ENV = {
    "RIIDO_DEVICE_SECRET", "RIIDO_SAAS_TOKEN", "RIIDO_TOKEN",
    "RIIDO_LEASE_TOKEN", "LEASE_TOKEN",
}
READ_CHECKS = ["context", "work", "parent", "tree", "relations", "comments",
               "agents", "outside_scope", "activity_unsupported", "search_unsupported",
               "changes_unsupported"]
MAX_RESPONSE_BYTES = 17 * 1024 * 1024


class ProbeFailure(Exception):
    """Only locally defined, non-sensitive reason codes may be reported."""


def require(condition, reason):
    if not condition:
        raise ProbeFailure(reason)


def envelope(result, expected_exit=0, expected_error=None):
    code, raw = result
    require(len(raw) <= MAX_RESPONSE_BYTES, "RESPONSE_TOO_LARGE")
    try:
        value = json.loads(raw)
    except (ValueError, UnicodeError):
        raise ProbeFailure("INVALID_JSON") from None
    require(isinstance(value, dict), "INVALID_ENVELOPE")
    require(value.get("schema_version") == "1", "SCHEMA_MISMATCH")
    require(isinstance(value.get("request_id"), str) and bool(value["request_id"]),
            "REQUEST_ID_MISSING")
    require(isinstance(value.get("sources"), list)
            and isinstance(value.get("warnings"), list), "ARRAY_CONTRACT")
    require("data" in value and "page" in value and "error" in value,
            "ENVELOPE_FIELDS_MISSING")
    require(code == expected_exit, "EXIT_MISMATCH")
    if expected_error:
        error = value["error"]
        require(isinstance(error, dict) and error.get("code") == expected_error,
                "ERROR_CODE_MISMATCH")
        require(value["data"] is None and not value["sources"], "FAILURE_DATA_LEAK")
        require(error.get("request_id") == value["request_id"], "ERROR_REQUEST_MISMATCH")
    else:
        require(value["error"] is None and isinstance(value["data"], dict),
                "SUCCESS_CONTRACT")
    return value


def snapshot_envelope(result):
    value = envelope(result)
    require(all(isinstance(w, dict) for w in value["warnings"]), "WARNING_CONTRACT")
    require("ASSIGNMENT_SNAPSHOT" in {w.get("code") for w in value["warnings"]},
            "SNAPSHOT_WARNING_MISSING")
    for source in value["sources"]:
        require(isinstance(source, dict), "SOURCE_CONTRACT")
        require(source.get("status") in {"available", "partial"}, "SOURCE_STATUS")
        require(isinstance(source.get("truncated"), bool), "SOURCE_TRUNCATION_FLAG")
        require(isinstance(source.get("fetched_at"), str) and bool(source["fetched_at"]),
                "SOURCE_CAPTURE_TIME_MISSING")
    return value


AGENT_FIELDS = {"agent_id", "name", "role_status", "work_status", "runtime", "callable"}
AGENT_PRIVATE_KEYS = {"instruction", "owner_principal_id", "model_id", "lease_token"}


def validate_agents(listing, shown):
    """RIID-5892 directory: live, requester-scoped, no private agent fields."""
    require("LIVE_STATUS" in {w.get("code") for w in listing["warnings"] if isinstance(w, dict)},
            "LIVE_WARNING_MISSING")
    agents = listing["data"].get("agents")
    require(isinstance(agents, list) and len(agents) >= 1, "AGENT_FIXTURE_MISSING")
    current = 0
    for entry in agents:
        require(isinstance(entry, dict) and AGENT_FIELDS <= entry.keys(), "AGENT_FIELDS_MISSING")
        require(not AGENT_PRIVATE_KEYS & entry.keys(), "AGENT_PRIVATE_FIELD_LEAK")
        require(isinstance(entry["callable"], bool)
                and entry["callable"] == (not entry.get("unavailable_reasons")), "AGENT_CALLABLE_MISMATCH")
        require(entry["runtime"].get("connection") in {"online", "offline"}, "AGENT_CONNECTION_INVALID")
        current += 1 if entry.get("is_current_agent") else 0
    require(current == 1, "CURRENT_AGENT_NOT_MARKED")
    require(listing["data"].get("policy_checked") is False or
            listing["data"].get("policy_checked") is True, "POLICY_FLAG_MISSING")
    if shown is not None:
        require(shown["data"].get("agent", {}).get("agent_id") in {a["agent_id"] for a in agents},
                "AGENT_SHOW_MISMATCH")
    return len(agents)


def validate_comments(pages, expected_total):
    seen, previous_priority, capture_times = set(), 0, set()
    for index, value in enumerate(pages):
        data, page = value["data"], value["page"]
        comments = data.get("comments")
        require(isinstance(comments, list), "COMMENTS_ARRAY")
        require(data.get("total") == expected_total, "COMMENT_TOTAL_MISMATCH")
        require(data.get("cursor") == str(index), "CURSOR_MISMATCH")
        require(isinstance(page, dict) and page.get("limit") == 20, "PAGE_LIMIT")
        next_cursor = str(index + 1) if (index + 1) * 20 < expected_total else ""
        require(data.get("next_cursor", "") == next_cursor
                and page.get("next_cursor", "") == next_cursor, "NEXT_CURSOR_MISMATCH")
        require(page.get("truncated") is bool(next_cursor), "PAGE_TRUNCATION_FLAG")
        require(len(comments) == min(20, expected_total - index * 20), "PAGE_COUNT_MISMATCH")
        source_ids = {s.get("id") for s in value["sources"] if s.get("kind") == "comment"}
        require(all(s.get("version") for s in value["sources"]), "COMMENT_VERSION_MISSING")
        page_ids = set()
        for comment in comments:
            require(isinstance(comment, dict), "COMMENT_CONTRACT")
            identifier, priority = comment.get("id"), comment.get("priority")
            require(isinstance(identifier, str) and bool(identifier), "COMMENT_ID_MISSING")
            require(identifier not in seen, "DUPLICATE_COMMENT")
            require(type(priority) is int and 1 <= priority <= 3
                    and priority >= previous_priority, "COMMENT_PRIORITY_ORDER")
            require(isinstance(comment.get("body"), str), "COMMENT_BODY_MISSING")
            previous_priority = priority
            seen.add(identifier)
            page_ids.add(identifier)
        require(source_ids == page_ids, "COMMENT_SOURCE_MISMATCH")
        capture_times.update(s["fetched_at"] for s in value["sources"])
    require(len(seen) == expected_total, "COMMENTS_INCOMPLETE")
    require(len(capture_times) == 1, "COMMENT_SNAPSHOT_CHANGED")
    return len(seen)


def new_report(mode):
    return {"report_schema": "1", "mode": mode, "staging_acceptance": "NOT_COMPLETE",
            "checks": [], "manual_checks": [{"name": n, "status": "NOT_RUN"}
                                            for n in MANUAL_CHECKS]}


def run_suite(runner, work, parent, outside, expected_comments):
    report = new_report("approved_agent_read_probe")
    captures = set()

    def read(args):
        value = snapshot_envelope(runner(args + ["--json"]))
        captures.update(s["fetched_at"] for s in value["sources"])
        require(len(captures) <= 1, "SNAPSHOT_CHANGED_DURING_PROBE")
        return value

    def check(name, fn):
        try:
            count = fn()
            item = {"name": name, "status": "PASS"}
            if type(count) is int:
                item["count"] = count
            report["checks"].append(item)
            return True
        except ProbeFailure as exc:
            report["checks"].append({"name": name, "status": "FAIL", "reason": str(exc)})
            return False

    def body(args, target):
        value = read(args)
        data = value["data"]
        identity = data.get("component", data)
        require(isinstance(identity, dict) and
                (identity.get("id") == target or str(identity.get("keyNumber", "")).lower() == target.lower()),
                "RETURNED_WORK_MISMATCH")
        document = data.get("document")
        require(isinstance(document, dict) and bool(document.get("content")), "FIXTURE_BODY_MISSING")
        require(any(s.get("id") == identity.get("id") and s.get("version") for s in value["sources"]),
                "BODY_VERSION_MISSING")
        return len(value["sources"])

    def tree():
        value = read(["work", "tree", work])
        hierarchy = value["data"].get("hierarchy")
        require(isinstance(hierarchy, dict), "HIERARCHY_MISSING")
        for ref in hierarchy.values():
            if isinstance(ref, dict):
                require(not ref.get("document") and not ref.get("resources"), "TREE_BODY_LEAK")
        return len(value["sources"])

    def relations():
        value = read(["work", "relations", work])
        related = value["data"].get("related")
        require(isinstance(related, dict), "RELATED_FIXTURE_MISSING")
        for kind in ["related_work", "related_meetings"]:
            entries = related.get(kind)
            require(isinstance(entries, list) and len(entries) >= 1, "RELATED_FIXTURE_MISSING")
            for entry in entries:
                require(isinstance(entry, dict) and bool(entry.get("title")), "RELATED_TITLE_MISSING")
                require(not any(key in entry for key in ["body", "document", "content"]),
                        "RELATED_BODY_LEAK")
        return len(value["sources"])

    def comments():
        pages, cursor = [], ""
        for _ in range(10):
            args = ["work", "comments", work] + (["--cursor", cursor] if cursor else [])
            value = read(args)
            pages.append(value)
            page = value["page"]
            require(isinstance(page, dict), "PAGE_MISSING")
            cursor = page.get("next_cursor", "")
            require(isinstance(cursor, str) and (not cursor or cursor.isdecimal()), "INVALID_CURSOR")
            if not cursor:
                return validate_comments(pages, expected_comments)
            require(cursor == str(len(pages)), "REPEATED_OR_SKIPPED_CURSOR")
        raise ProbeFailure("PAGE_BUDGET_EXCEEDED")

    def agents():
        listing = envelope(runner(["agents", "list", "--json"]))
        others = [a.get("agent_id") for a in listing["data"].get("agents") or []
                  if isinstance(a, dict) and not a.get("is_current_agent")]
        shown = envelope(runner(["agents", "show", others[0], "--json"])) if others else None
        return validate_agents(listing, shown)

    checks = [
        ("context", lambda: body(["context", "get"], work)),
        ("work", lambda: body(["work", "get", work], work)),
        ("parent", lambda: body(["work", "get", parent], parent)),
        ("tree", tree), ("relations", relations), ("comments", comments), ("agents", agents),
        ("outside_scope", lambda: envelope(runner(["work", "get", outside, "--json"]),
                                           4, "OUT_OF_RUN_SCOPE")),
        ("activity_unsupported", lambda: envelope(runner(["work", "activity", work, "--json"]),
                                                  8, "CONTROL_PLANE_UNSUPPORTED")),
        ("search_unsupported", lambda: envelope(runner(["work", "search", "--query", "qa", "--json"]),
                                                8, "CONTROL_PLANE_UNSUPPORTED")),
        ("changes_unsupported", lambda: envelope(runner(["context", "changes", "--since", "qa", "--json"]),
                                                 8, "CONTROL_PLANE_UNSUPPORTED")),
    ]
    for index, (name, fn) in enumerate(checks):
        if not check(name, fn):
            report["checks"].extend({"name": n, "status": "NOT_RUN"} for n, _ in checks[index + 1:])
            break
    report["automated_checks"] = "PASS" if all(x["status"] == "PASS" for x in report["checks"]) else "FAIL"
    return report


def live_runner(binary, environ):
    token = environ["RIIDO_RUN_TOKEN"].encode()

    def run(args):
        try:
            result = subprocess.run([binary] + args, env=environ, capture_output=True,
                                    timeout=25, check=False)
        except subprocess.TimeoutExpired:
            raise ProbeFailure("COMMAND_TIMEOUT") from None
        except OSError:
            raise ProbeFailure("COMMAND_START_FAILED") from None
        require(token not in result.stdout and token not in result.stderr, "CAPABILITY_OUTPUT_LEAK")
        require(not result.stderr.strip(), "UNEXPECTED_STDERR")
        return result.returncode, result.stdout
    return run


def preflight(args, environ):
    require(bool(args.staging_approval) and bool(args.daemon_build), "APPROVAL_REFERENCES_REQUIRED")
    require(bool(re.fullmatch(r"[0-9a-f]{64}", args.cli_sha256 or "")), "CLI_HASH_REQUIRED")
    require(all(environ.get(key, "").strip() for key in
                ["RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID"]), "AGENT_ENV_REQUIRED")
    require(not any(k.upper() in FORBIDDEN_ENV and v for k, v in environ.items()),
            "PRIVILEGED_CREDENTIAL_IN_AGENT_ENV")
    require(25 <= args.expected_comments <= 200, "FIXTURE_COMMENT_COUNT")
    require(all(re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]{0,127}", ref or "")
                for ref in [args.work, args.parent, args.outside]), "FIXTURE_REFS_REQUIRED")
    require(len({args.work, args.parent, args.outside}) == 3, "FIXTURE_REFS_MUST_DIFFER")
    binary = shutil.which("riido", path=environ.get("PATH"))
    require(bool(binary), "CLI_NOT_ON_PATH")
    try:
        actual_hash = hashlib.sha256(Path(binary).read_bytes()).hexdigest()
    except OSError:
        raise ProbeFailure("CLI_UNREADABLE") from None
    require(actual_hash == args.cli_sha256, "CLI_HASH_MISMATCH")
    return binary


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--execute", action="store_true", help="Only after recorded staging approvals")
    parser.add_argument("--staging-approval", help="Non-secret decision/approval reference, not a credential")
    parser.add_argument("--daemon-build", help="Approved build reference; operator must verify running daemon")
    parser.add_argument("--cli-sha256", help="Expected SHA-256 from the approved build, not an arbitrary local hash")
    parser.add_argument("--work", help="Assigned synthetic fixture work")
    parser.add_argument("--parent", help="In-snapshot hierarchy item with nonempty body")
    parser.add_argument("--outside", help="Work outside this assignment snapshot")
    parser.add_argument("--expected-comments", type=int, default=25)
    args = parser.parse_args(argv)
    if not args.execute:
        report = new_report("offline_plan")
        report["checks"] = [{"name": name, "status": "NOT_RUN"} for name in READ_CHECKS]
        exit_code = 0
    else:
        try:
            binary = preflight(args, os.environ)
            report = run_suite(live_runner(binary, dict(os.environ)), args.work, args.parent,
                               args.outside, args.expected_comments)
            exit_code = 0 if report["automated_checks"] == "PASS" else 1
        except ProbeFailure as exc:
            report = new_report("blocked_preflight")
            report["checks"] = [{"name": "preflight", "status": "FAIL", "reason": str(exc)}]
            exit_code = 2
    print(json.dumps(report, indent=2))
    return exit_code


if __name__ == "__main__":
    sys.exit(main())
