#!/bin/bash
set -euo pipefail
helper_dir="$(cd "$(dirname "$0")" && pwd)"
fixture_dir="$(mktemp -d "${TMPDIR:-/tmp}/riido-cli-auth-fixture.XXXXXX")"
trap 'rm -rf "$fixture_dir"' EXIT
swiftc -warnings-as-errors -D CLI_AUTH_HELPER_TESTING "$helper_dir"/Sources/*.swift "$helper_dir"/Tests/*.swift -o "$fixture_dir/fixtures"
"$fixture_dir/fixtures" "$helper_dir/Tests/testdata/installation-proof-v1.json" ${RIIDO_INSTALLATION_PROOF_FIXTURE_OUTPUT:+"$RIIDO_INSTALLATION_PROOF_FIXTURE_OUTPUT"}
bundle="$($helper_dir/build.sh "$fixture_dir")"
request='{"version":1,"operation":"status","key":"fixture_account_key_5900","interactive":false}'
response="$(printf '%s' "$request" | "$bundle/Contents/MacOS/riido-cli-auth-helper")"
confirm_request='{"version":1,"operation":"confirm","key":"fixture_account_key_5900","interactive":true}'
confirm_response="$(printf '%s' "$confirm_request" | "$bundle/Contents/MacOS/riido-cli-auth-helper")"
python3 - "$response" "$confirm_response" "$helper_dir" "$bundle/Contents/MacOS/riido-cli-auth-helper" <<'PYTHON'
import base64,json,subprocess,sys,time
responses = sys.argv[1:3]
vector = json.load(open(sys.argv[3]+'/Tests/testdata/installation-proof-v1.json'))
scope = {'resource':vector['Resource'], 'client_id':vector['ClientID']}
proof = dict(scope, installation_intent=vector['IntentID'], challenge_hash=vector['ChallengeHash'],
             key_fingerprint=vector['KeyFingerprint'], code=vector['Code'], redirect_uri=vector['RedirectURI'],
             code_challenge=vector['CodeChallenge'], state=vector['State'], expires_at=int(time.time())+300)
for operation, payload in [('installation-key',scope),('installation-proof',proof)]:
    request = {'version':1,'operation':operation,'key':'08eba32e93e89b49cbed7c040f8de7dba724391c4015de3baaa78d9dad03957f',
               'interactive':True,'value':base64.b64encode(json.dumps(payload).encode()).decode()}
    responses.append(subprocess.run([sys.argv[4]], input=json.dumps(request),text=True,capture_output=True,check=True,timeout=10).stdout)
for body in responses:
    response=json.loads(body)
    assert response['version']==1 and response['ok'] is False
    assert response['error']['code']=='UNAVAILABLE'
    assert 'data' not in response
print('PASS: unsigned helper refuses credential/confirmation/installation key/proof before keychain or OS access')
PYTHON
