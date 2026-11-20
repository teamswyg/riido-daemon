#!/usr/bin/env python3
"""Release-only profile/signing/startup gate. Never reads credential items."""
import argparse
import datetime
import hashlib
from pathlib import Path
import plistlib
import re
import subprocess
import sys
import tempfile

TEAM = "3PVV8DQPL6"
APP = TEAM + ".io.riido.cli-auth-helper"
MARKER = b"riido-cli-helper-preflight.v1\n"


def run(command, **kwargs):
    # Child diagnostics may contain profile data; emit only our fixed errors.
    try:
        return subprocess.run(command, capture_output=True, timeout=15, check=True, **kwargs).stdout
    except (OSError, subprocess.SubprocessError) as error:
        raise ValueError("profile/signature/startup command failed") from error


def profile_certificates(profile, now=None):
    now = now or datetime.datetime.now(datetime.timezone.utc)
    expiry = profile.get("ExpirationDate")
    if not isinstance(expiry, datetime.datetime):
        raise ValueError("profile has no expiry")
    if expiry.replace(tzinfo=datetime.timezone.utc) <= now:
        raise ValueError("profile expired")
    entitlements = profile.get("Entitlements", {})
    if (profile.get("TeamIdentifier") != [TEAM]
            or entitlements.get("com.apple.application-identifier") != APP
            or not any(group in (APP, TEAM + ".*") for group in entitlements.get("keychain-access-groups", []))
            or profile.get("ProvisionsAllDevices") is not True):
        raise ValueError("profile does not authorize the distributed helper")
    certificates = profile.get("DeveloperCertificates", [])
    if not certificates or any(not isinstance(cert, bytes) or not cert for cert in certificates):
        raise ValueError("profile has no authorized certificates")
    return certificates


def load_profile(path):
    try:
        profile = plistlib.loads(run(["/usr/bin/security", "cms", "-D", "-i", str(path)]))
    except (plistlib.InvalidFileException, TypeError) as error:
        raise ValueError("invalid provisioning profile") from error
    return profile_certificates(profile)


def select_identity(certificates, listing, requested):
    allowed = {hashlib.sha1(cert).hexdigest().upper() for cert in certificates}
    matches = set()
    for fingerprint, name in re.findall(r'\b([A-Fa-f0-9]{40})\s+"([^"\n]+)"', listing):
        fingerprint = fingerprint.upper()
        if (fingerprint in allowed and name.startswith("Developer ID Application:")
                and requested in (name, fingerprint, fingerprint.lower())):
            matches.add(fingerprint)
    if len(matches) != 1:
        raise ValueError("expected exactly one usable requested identity authorized by the profile")
    return matches.pop()


def verify_signer(certificates, leaf):
    if leaf not in certificates:
        raise ValueError("helper signing certificate is not authorized by embedded profile")


def verify_startup(executable):
    # Empty stdin and the dedicated flag never invoke Keychain/user-presence APIs.
    output = run([str(executable), "--preflight"], input=b"")
    if output != MARKER:
        raise ValueError("helper did not complete signing-only startup preflight")


def verify_bundle(bundle):
    certificates = load_profile(bundle / "Contents/embedded.provisionprofile")
    with tempfile.TemporaryDirectory(prefix="riido-helper-public-cert-") as directory:
        prefix = str(Path(directory) / "cert-")
        run(["/usr/bin/codesign", "-d", "--extract-certificates=" + prefix, str(bundle)])
        verify_signer(certificates, Path(prefix + "0").read_bytes())
    verify_startup(bundle / "Contents/MacOS/riido-cli-auth-helper")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    select = commands.add_parser("select-identity")
    select.add_argument("--profile", type=Path, required=True)
    select.add_argument("--identity", required=True)
    verify = commands.add_parser("verify-bundle")
    verify.add_argument("--bundle", type=Path, required=True)
    args = parser.parse_args()
    try:
        if args.command == "select-identity":
            certificates = load_profile(args.profile)
            identities = run(["/usr/bin/security", "find-identity", "-v", "-p", "codesigning"]).decode()
            print(select_identity(certificates, identities, args.identity))
        else:
            verify_bundle(args.bundle)
    except (ValueError, OSError) as error:
        print("CLI helper release gate: " + str(error), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
