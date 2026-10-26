"""Privacy and Unix boundary tests; all credentials are synthetic."""

import contextlib
import io
import json
import os
from pathlib import Path
import socket
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import ended_run_ipc as ipc
import ended_run_probe as probe


class EndedRunIPCTests(unittest.TestCase):
    def test_private_socket_and_peer(self):
        with tempfile.TemporaryDirectory() as raw:
            directory = Path(raw).resolve()
            os.chmod(directory, 0o700)
            path = str(directory / "handoff.sock")
            self.assertEqual(ipc.private_path(path), path)
            with socket.socket(socket.AF_UNIX) as listener:
                listener.bind(path)
                os.chmod(path, 0o600)
                listener.listen(1)
                with socket.socket(socket.AF_UNIX) as client:
                    client.connect(path)
                    server, _ = listener.accept()
                    with server:
                        ipc.same_user_peer(client)
                        ipc.same_user_peer(server)
                        self.assertEqual(ipc.private_path(path, True), path)
                        ipc.send_json(client, {"accepted": True})
                        self.assertEqual(ipc.read_json(server), {"accepted": True})

    def test_open_directory_or_socket_rejected(self):
        with tempfile.TemporaryDirectory() as raw:
            directory = Path(raw).resolve()
            path = str(directory / "handoff.sock")
            os.chmod(directory, 0o755)
            with self.assertRaisesRegex(ipc.Failure, "HANDOFF_PARENT_NOT_PRIVATE"):
                ipc.private_path(path)
            os.chmod(directory, 0o700)
            with socket.socket(socket.AF_UNIX) as listener:
                listener.bind(path)
                os.chmod(path, 0o666)
                with self.assertRaisesRegex(ipc.Failure, "HANDOFF_SOCKET_NOT_PRIVATE"):
                    ipc.private_path(path, True)

    def test_symlink_parent_rejected(self):
        with tempfile.TemporaryDirectory() as raw:
            directory = Path(raw).resolve()
            link = directory / "link"
            link.symlink_to(directory, target_is_directory=True)
            with self.assertRaises(ipc.Failure):
                ipc.private_path(str(link / "handoff.sock"))

    def test_malformed_ipc_never_echoes_payload(self):
        one, two = socket.socketpair()
        with one, two:
            one.sendall(b"private-fixture-value\n")
            with self.assertRaisesRegex(ipc.Failure, "^IPC_INVALID_JSON$"):
                ipc.read_json(two)

    def test_byte_trickle_cannot_extend_absolute_read_budget(self):
        class Trickle:
            now = 0
            def gettimeout(self): return 0.05
            def settimeout(self, value): self.limit = value
            def recv(self, _):
                self.now += 0.03
                return b" "
        conn = Trickle()
        with patch.object(ipc.time, "monotonic", side_effect=lambda: conn.now):
            with self.assertRaisesRegex(ipc.Failure, "^PROBE_TIMEOUT$"):
                ipc.read_json(conn)
        self.assertLessEqual(conn.now, 0.06)

    def test_cli_keeps_capability_out_of_argv_and_reports(self):
        payload = dict(zip(probe.FIELDS, ["rrc_fixture_private_capability", "/fixture", "task", "workspace"]))
        output = json.dumps({"schema_version": "1", "data": None, "error": {"code": "PERMISSION_DENIED"}}).encode()
        with patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 4, output, b"")) as run:
            probe.cli_runner("/fixture/riido", payload)(["context", "get", "--json"])
            self.assertNotIn(payload["RIIDO_RUN_TOKEN"], str(run.call_args.args))
            self.assertEqual(run.call_args.kwargs["env"]["RIIDO_RUN_TOKEN"], payload["RIIDO_RUN_TOKEN"])
        with patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 0, payload["RIIDO_RUN_TOKEN"].encode(), b"")):
            with self.assertRaisesRegex(ipc.Failure, "^CAPABILITY_OUTPUT_LEAK$"):
                probe.cli_runner("/fixture/riido", payload)(["whoami", "--json"])

    def test_default_watch_and_missing_approval_make_no_connections(self):
        output = io.StringIO()
        with patch.object(socket, "socket") as create, contextlib.redirect_stdout(output):
            self.assertEqual(probe.main(["watch"]), 0)
            self.assertEqual(probe.main(["watch", "--execute", "--handoff", "/fixture"]), 1)
            create.assert_not_called()
        self.assertIn("NOT_RUN", output.getvalue())
        self.assertIn("APPROVAL_REFERENCES_REQUIRED", output.getvalue())

    def test_sensitive_environment_and_raw_failure_are_closed(self):
        with self.assertRaisesRegex(ipc.Failure, "^PRIVILEGED_CREDENTIAL_PRESENT$"):
            probe.safe_environment({"riido_device_secret": "private-fixture-value"})
        output = io.StringIO()
        with patch.object(probe, "send", side_effect=OSError("private-fixture-value")), contextlib.redirect_stdout(output):
            self.assertEqual(probe.main(["send", "--handoff", "/fixture"]), 1)
        self.assertNotIn("private-fixture-value", output.getvalue())
        self.assertIn("LOCAL_PROBE_ERROR", output.getvalue())


if __name__ == "__main__":
    unittest.main()
