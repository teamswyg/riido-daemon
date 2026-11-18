import datetime
import hashlib
import importlib.util
from pathlib import Path
import subprocess
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("gate", Path(__file__).with_name("verify-macos-helper.py"))
gate = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gate)


class HelperReleaseGateTests(unittest.TestCase):
    def profile(self):
        return {"ExpirationDate": datetime.datetime(2040, 1, 1), "TeamIdentifier": [gate.TEAM],
                "ProvisionsAllDevices": True, "DeveloperCertificates": [b"authorized certificate"],
                "Entitlements": {"com.apple.application-identifier": gate.APP,
                                 "keychain-access-groups": [gate.TEAM + ".*"]}}

    def test_valid_profile(self):
        self.assertEqual(gate.profile_certificates(self.profile()), [b"authorized certificate"])

    def test_invalid_profile_boundaries(self):
        for key, value in [("ExpirationDate", datetime.datetime(2000, 1, 1)),
                           ("TeamIdentifier", ["other-team"]), ("ProvisionsAllDevices", False),
                           ("DeveloperCertificates", [])]:
            with self.subTest(key=key), self.assertRaises(ValueError):
                profile = self.profile(); profile[key] = value
                gate.profile_certificates(profile)
        for key, value in [("com.apple.application-identifier", "other-app"),
                           ("keychain-access-groups", ["other-team.*"])]:
            with self.subTest(key=key), self.assertRaises(ValueError):
                profile = self.profile(); profile["Entitlements"][key] = value
                gate.profile_certificates(profile)

    def test_shared_display_name_selects_only_profile_certificate(self):
        name = "Developer ID Application: Fixture (3PVV8DQPL6)"
        first = hashlib.sha1(b"first").hexdigest().upper()
        second = hashlib.sha1(b"second").hexdigest().upper()
        listing = f'1) {first} "{name}"\n2) {second} "{name}"'
        self.assertEqual(gate.select_identity([b"first"], listing, name), first)
        for certificates, request in [([b"third"], name), ([b"first", b"second"], name), ([b"first"], second)]:
            with self.subTest(request=request), self.assertRaises(ValueError):
                gate.select_identity(certificates, listing, request)

    def test_profile_mismatch_rejects_same_team_signer(self):
        gate.verify_signer([b"authorized certificate"], b"authorized certificate")
        with self.assertRaises(ValueError):
            gate.verify_signer([b"authorized certificate"], b"same-team different certificate")

    def test_startup_uses_only_dedicated_flag_and_empty_stdin(self):
        with patch.object(subprocess, "run", return_value=subprocess.CompletedProcess([], 0, gate.MARKER, b"")) as run:
            gate.verify_startup(Path("fixture-helper"))
            run.assert_called_once_with(["fixture-helper", "--preflight"], input=b"", capture_output=True, timeout=15, check=True)

    def test_legacy_zero_exit_is_not_success(self):
        for output in [b"", b'{"ok":false}', gate.MARKER + b"extra"]:
            with self.subTest(output=output), patch.object(gate, "run", return_value=output), self.assertRaises(ValueError):
                gate.verify_startup(Path("fixture-helper"))

    def test_crash_timeout_and_missing_executable_fail_closed(self):
        for error in [subprocess.CalledProcessError(-9, []), subprocess.TimeoutExpired([], 15), FileNotFoundError()]:
            with self.subTest(error=type(error)), patch.object(subprocess, "run", side_effect=error), self.assertRaisesRegex(ValueError, "command failed"):
                gate.verify_startup(Path("fixture-helper"))

    def test_mismatch_never_launches_helper(self):
        with patch.object(gate, "load_profile", return_value=[b"authorized"]), patch.object(gate, "run"), patch.object(Path, "read_bytes", return_value=b"wrong"), patch.object(gate, "verify_startup") as startup:
            with self.assertRaises(ValueError):
                gate.verify_bundle(Path("fixture.app"))
            startup.assert_not_called()


if __name__ == "__main__":
    unittest.main()
