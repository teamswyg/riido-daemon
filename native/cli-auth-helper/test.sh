#!/bin/bash
set -euo pipefail
helper_dir="$(cd "$(dirname "$0")" && pwd)"
fixture_dir="$(mktemp -d "${TMPDIR:-/tmp}/riido-cli-auth-fixture.XXXXXX")"
trap 'rm -rf "$fixture_dir"' EXIT
swiftc -warnings-as-errors -D CLI_AUTH_HELPER_TESTING "$helper_dir"/Sources/*.swift "$helper_dir/Tests/main.swift" -o "$fixture_dir/fixtures"
"$fixture_dir/fixtures"
bundle="$($helper_dir/build.sh "$fixture_dir")"
request='{"version":1,"operation":"status","key":"fixture_account_key_5900","interactive":false}'
response="$(printf '%s' "$request" | "$bundle/Contents/MacOS/riido-cli-auth-helper")"
confirm_request='{"version":1,"operation":"confirm","key":"fixture_account_key_5900","interactive":true}'
confirm_response="$(printf '%s' "$confirm_request" | "$bundle/Contents/MacOS/riido-cli-auth-helper")"
python3 - "$response" "$confirm_response" <<'PYTHON'
import json,sys
for body in sys.argv[1:]:
    response=json.loads(body)
    assert response['version']==1 and response['ok'] is False
    assert response['error']['code']=='UNAVAILABLE'
    assert 'data' not in response
print('PASS: unsigned helper refuses before keychain or OS confirmation access')
PYTHON
