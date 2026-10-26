"""Offline contract fixtures: no network, daemon, real token, or subprocess."""

import argparse
import hashlib
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import staging_probe as probe


SENSITIVE = "fixture-private-body-do-not-record"
CAPABILITY = "fixture-run-capability-do-not-record"


def success(data, sources=None, page=None):
    return {"schema_version": "1", "request_id": "cli_fixture", "data": data,
            "sources": sources if sources is not None else [source("task", "assigned")],
            "warnings": [{"code": "ASSIGNMENT_SNAPSHOT", "message": SENSITIVE}],
            "page": page, "error": None}


def source(kind, identifier):
    return {"kind": kind, "id": identifier, "version": "fixture-version",
            "fetched_at": "2026-10-26T00:00:00Z", "status": "available", "truncated": False}


def failure(code):
    value = success(None, [])
    value["warnings"] = []
    value["error"] = {"code": code, "message": SENSITIVE, "request_id": "cli_fixture",
                      "retryable": False}
    return value


def fixtures():
    out = {}
    for command in [("context", "get"), ("work", "get", "assigned")]:
        out[command] = (0, success({"component": {"id": "assigned"}, "document": {"content": SENSITIVE}}))
    out[("work", "get", "parent")] = (0, success({"id": "parent", "document": {"content": SENSITIVE}},
                                                 [source("project", "parent")]))
    out[("work", "tree", "assigned")] = (0, success({"hierarchy": {"project": {"title": SENSITIVE}}}))
    out[("work", "relations", "assigned")] = (0, success({"related": {
        "related_work": [{"id": "r1", "title": SENSITIVE}],
        "related_meetings": [{"id": "m1", "title": SENSITIVE}]}}))
    for index in range(2):
        comments = [{"id": f"c{i}", "body": SENSITIVE, "priority": 1 if i == 0 else 2 if i < 20 else 3}
                    for i in range(index * 20, min(25, (index + 1) * 20))]
        next_cursor = "1" if index == 0 else ""
        data = {"comments": comments, "total": 25, "cursor": str(index), "next_cursor": next_cursor}
        page = {"limit": 20, "next_cursor": next_cursor, "truncated": bool(next_cursor)}
        args = ("work", "comments", "assigned") + (("--cursor", "1") if index else ())
        out[args] = (0, success(data, [source("comment", c["id"]) for c in comments], page))
    live = {"code": "LIVE_STATUS", "message": SENSITIVE}
    agents = [{"agent_id": "self", "name": SENSITIVE, "role_status": "described", "work_status": "running",
               "runtime": {"connection": "online"}, "callable": False, "unavailable_reasons": ["self"],
               "is_current_agent": True},
              {"agent_id": "peer", "name": SENSITIVE, "role_status": "missing", "work_status": "idle",
               "runtime": {"connection": "offline"}, "callable": False, "unavailable_reasons": ["runtime_offline"]}]
    listing = success({"agents": agents, "policy_checked": False}, [source("agent", "peer")])
    listing["warnings"] = [live]
    out[("agents", "list")] = (0, listing)
    shown = success({"agent": agents[1], "policy_checked": False}, [source("agent", "peer")])
    shown["warnings"] = [live]
    out[("agents", "show", "peer")] = (0, shown)
    out[("work", "get", "outside")] = (4, failure("OUT_OF_RUN_SCOPE"))
    for command in [("work", "activity", "assigned"), ("work", "search", "--query", "qa"),
                    ("context", "changes", "--since", "qa")]:
        out[command] = (8, failure("CONTROL_PLANE_UNSUPPORTED"))
    return out


class ProbeTests(unittest.TestCase):
    def run_fixtures(self, values):
        self.calls = []

        def runner(args):
            self.assertEqual(args[-1], "--json")
            self.calls.append(tuple(args[:-1]))
            status, value = values[tuple(args[:-1])]
            return status, json.dumps(value).encode()
        return probe.run_suite(runner, "assigned", "parent", "outside", 25)

    def test_current_contract_and_no_sensitive_evidence(self):
        report = self.run_fixtures(fixtures())
        self.assertEqual(report["automated_checks"], "PASS")
        self.assertEqual(report["staging_acceptance"], "NOT_COMPLETE")
        self.assertTrue(all(c["status"] == "NOT_RUN" for c in report["manual_checks"]))
        self.assertEqual(len(self.calls), 13)
        self.assertNotIn(SENSITIVE, json.dumps(report))
        self.assertNotIn("assigned", json.dumps(report))

    def test_duplicate_page_fails_and_stops_further_requests(self):
        values = fixtures()
        values[("work", "comments", "assigned", "--cursor", "1")][1]["data"]["comments"][0]["id"] = "c0"
        report = self.run_fixtures(values)
        self.assertEqual(report["checks"][5]["reason"], "DUPLICATE_COMMENT")
        self.assertNotIn(("work", "get", "outside"), self.calls)

    def test_cursor_loop_stops_at_first_page(self):
        values = fixtures()
        values[("work", "comments", "assigned")][1]["page"]["next_cursor"] = "0"
        report = self.run_fixtures(values)
        self.assertEqual(report["checks"][5]["reason"], "REPEATED_OR_SKIPPED_CURSOR")
        self.assertEqual(len(self.calls), 6)

    def test_partial_fixture_does_not_pass_as_complete(self):
        values = fixtures()
        values[("work", "comments", "assigned", "--cursor", "1")][1]["data"]["comments"].pop()
        report = self.run_fixtures(values)
        self.assertEqual(report["checks"][5]["reason"], "PAGE_COUNT_MISMATCH")

    def test_related_body_leak_rejected(self):
        values = fixtures()
        values[("work", "relations", "assigned")][1]["data"]["related"]["related_work"][0]["body"] = SENSITIVE
        report = self.run_fixtures(values)
        self.assertEqual(report["checks"][4]["reason"], "RELATED_BODY_LEAK")
        self.assertNotIn(SENSITIVE, json.dumps(report))

    def test_parent_read_cannot_silently_return_assigned_task(self):
        values = fixtures()
        values[("work", "get", "parent")] = values[("work", "get", "assigned")]
        report = self.run_fixtures(values)
        self.assertEqual(report["checks"][2]["reason"], "RETURNED_WORK_MISMATCH")

    def test_snapshot_warning_and_capture_consistency(self):
        values = fixtures()
        values[("context", "get")][1]["warnings"] = []
        self.assertEqual(self.run_fixtures(values)["checks"][0]["reason"], "SNAPSHOT_WARNING_MISSING")
        values = fixtures()
        values[("work", "get", "assigned")][1]["sources"][0]["fetched_at"] = "2026-10-26T01:00:00Z"
        self.assertEqual(self.run_fixtures(values)["checks"][1]["reason"], "SNAPSHOT_CHANGED_DURING_PROBE")

    def test_expected_denial_is_specific(self):
        values = fixtures()
        values[("work", "get", "outside")] = (4, failure("PERMISSION_DENIED"))
        self.assertEqual(self.run_fixtures(values)["checks"][7]["reason"], "ERROR_CODE_MISMATCH")

    def test_non_json_failure_never_echoes_output(self):
        with self.assertRaisesRegex(probe.ProbeFailure, "^INVALID_JSON$"):
            probe.envelope((0, SENSITIVE.encode()))

    def test_secret_output_and_timeout_never_echo_output(self):
        runner = probe.live_runner("fixture-cli", {"RIIDO_RUN_TOKEN": CAPABILITY})
        with patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 0, CAPABILITY.encode(), b"")):
            with self.assertRaisesRegex(probe.ProbeFailure, "^CAPABILITY_OUTPUT_LEAK$"):
                runner(["context", "get", "--json"])
        with patch.object(subprocess, "run", side_effect=subprocess.TimeoutExpired("fixture-cli", 25, output=SENSITIVE)):
            with self.assertRaisesRegex(probe.ProbeFailure, "^COMMAND_TIMEOUT$"):
                runner(["context", "get", "--json"])

    def test_preflight_fails_before_any_command(self):
        args = argparse.Namespace(staging_approval=None, daemon_build=None, cli_sha256=None,
                                  expected_comments=25, work="assigned", parent="parent", outside="outside")
        with patch.object(subprocess, "run") as run:
            with self.assertRaisesRegex(probe.ProbeFailure, "APPROVAL_REFERENCES_REQUIRED"):
                probe.preflight(args, {})
            run.assert_not_called()

    def test_preflight_credential_case_and_binary_hash(self):
        with tempfile.TemporaryDirectory() as directory:
            binary = Path(directory) / "riido"
            binary.write_bytes(b"offline-fixture-never-executed")
            args = argparse.Namespace(staging_approval="fixture", daemon_build="fixture",
                                      cli_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                                      expected_comments=25, work="assigned", parent="parent", outside="outside")
            env = {"RIIDO_RUN_TOKEN": CAPABILITY, "RIIDO_DAEMON_SOCKET": "fixture", "RIIDO_TASK_ID": "assigned"}
            with patch.object(probe.shutil, "which", return_value=str(binary)):
                self.assertEqual(probe.preflight(args, env), str(binary))
                with self.assertRaisesRegex(probe.ProbeFailure, "PRIVILEGED_CREDENTIAL_IN_AGENT_ENV"):
                    probe.preflight(args, dict(env, riido_device_secret=SENSITIVE))
                args.cli_sha256 = "0" * 64
                with self.assertRaisesRegex(probe.ProbeFailure, "CLI_HASH_MISMATCH"):
                    probe.preflight(args, env)


if __name__ == "__main__":
    unittest.main()


class AgentDirectoryTests(unittest.TestCase):
    def run_agents(self, mutate):
        values = fixtures()
        mutate(values)

        def runner(args):
            status, value = values[tuple(args[:-1])]
            return status, json.dumps(value).encode()
        report = probe.run_suite(runner, "assigned", "parent", "outside", 25)
        return next(c for c in report["checks"] if c["name"] == "agents")

    def test_private_fields_and_snapshot_warning_fail(self):
        def leak(values):
            values[("agents", "list")][1]["data"]["agents"][1]["instruction"] = SENSITIVE
        self.assertEqual(self.run_agents(leak)["reason"], "AGENT_PRIVATE_FIELD_LEAK")

        def stale(values):
            values[("agents", "list")][1]["warnings"] = [{"code": "ASSIGNMENT_SNAPSHOT"}]
        self.assertEqual(self.run_agents(stale)["reason"], "LIVE_WARNING_MISSING")

        def inconsistent(values):
            values[("agents", "list")][1]["data"]["agents"][1]["callable"] = True
        self.assertEqual(self.run_agents(inconsistent)["reason"], "AGENT_CALLABLE_MISMATCH")
