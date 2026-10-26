"""Offline ended-run probe regressions. No provider or staging calls."""

import copy
import json
import unittest

import ended_run_probe as probe


TOKEN = "rrc_fixture_private_capability"
BINDING = {"assignment_id": "assignment", "task_id": "task", "workspace_id": "workspace", "runtime_id": "runtime"}


def status(active=True, pid=101):
    return {"schema_version": "riido-agent-daemon-status.v1", "pid": pid,
            "started_at": "2026-10-27T00:00:00Z", "daemon_version": "qa",
            "ready": True, "metrics": {"runtime_count": 1, "runtime_responding": 1},
            "runtimes": [{"runtime_id": "runtime", "health": "healthy",
                          "running_tasks": [{"task_id": "assignment"}] if active else []}],
            "executions": [dict(BINDING, phase="running")] if active else []}


def envelope(data=None, code=None):
    return {"schema_version": "1", "request_id": "fixture", "data": data,
            "sources": [], "warnings": [], "page": None,
            "error": {"code": code, "request_id": "fixture"} if code else None}


class EndedRunTests(unittest.TestCase):
    def run_case(self, observations, last=None):
        calls = []
        reports = iter([*observations, observations[-1]])
        clock = [0]

        def cli(args):
            calls.append(args)
            if len(calls) == 1:
                return 0, envelope({"mode": "agent", "run": BINDING})
            return last or (4, envelope(code="PERMISSION_DENIED"))

        def sleep(seconds):
            clock[0] += seconds

        result = probe.observe_ended_run(cli, lambda: next(reports), "task", "workspace",
                                        5, clock=lambda: clock[0], sleep=sleep)
        return result, calls

    def test_active_to_ended_reuses_once_and_never_exposes_identity(self):
        report, calls = self.run_case([status(), status(False), status(False)])
        self.assertEqual(report["status"], "PASS")
        self.assertEqual(calls, [["whoami", "--json"], ["whoami", "--json"], ["context", "get", "--json"]])
        self.assertEqual(report["staging_acceptance"], "NOT_COMPLETE")
        self.assertNotIn("assignment", json.dumps(report))
        self.assertNotIn(TOKEN, json.dumps(report))

    def test_daemon_restart_is_not_capability_expiry(self):
        with self.assertRaisesRegex(probe.Failure, "^DAEMON_CHANGED$"):
            self.run_case([status(), status(False, pid=102)])

    def test_no_active_baseline_cannot_pass(self):
        with self.assertRaisesRegex(probe.Failure, "^ACTIVE_EXECUTION_NOT_OBSERVED$"):
            self.run_case([status(False)])

    def test_nonresponsive_runtime_cannot_count_as_ended(self):
        broken = status(False)
        broken["metrics"]["runtime_responding"] = 0
        with self.assertRaisesRegex(probe.Failure, "^STATUS_INCONCLUSIVE$"):
            self.run_case([status(), broken])

    def test_other_execution_cannot_supply_baseline(self):
        wrong = status()
        wrong["executions"][0]["assignment_id"] = "other"
        with self.assertRaisesRegex(probe.Failure, "^ACTIVE_EXECUTION_NOT_OBSERVED$"):
            self.run_case([wrong])

    def test_pending_native_slot_does_not_count_as_ended(self):
        native = status(False)
        native["runtimes"][0]["running_tasks"] = [{"task_id": "assignment"}]
        report, _ = self.run_case([status(), native, status(False), status(False)])
        self.assertEqual(report["status"], "PASS")
        self.assertEqual(report["status_observations"], 4)

    def test_denial_must_be_specific_with_empty_data(self):
        for last in [(10, envelope(code="LEASE_LOST")), (4, envelope(code="OUT_OF_RUN_SCOPE")),
                     (0, envelope({"private": "body"})),
                     (4, envelope({"private": "body"}, "PERMISSION_DENIED"))]:
            with self.subTest(last=last):
                with self.assertRaises(probe.Failure):
                    self.run_case([status(), status(False), status(False)], last)

    def test_snapshot_values_are_not_mutated(self):
        original = status()
        saved = copy.deepcopy(original)
        self.run_case([original, status(False), status(False)])
        self.assertEqual(original, saved)

    def test_server_permission_denial_is_not_local_revocation(self):
        reports = iter([status(), status(False), status(False)])
        def broken_cli(args):
            if args == ["whoami", "--json"]:
                return 0, envelope({"mode": "agent", "run": BINDING})
            return 4, envelope(code="PERMISSION_DENIED")
        with self.assertRaisesRegex(probe.Failure, "^ENDED_CAPABILITY_NOT_DENIED$"):
            probe.observe_ended_run(broken_cli, lambda: next(reports), "task", "workspace", 5, sleep=lambda _: None)


if __name__ == "__main__":
    unittest.main()
