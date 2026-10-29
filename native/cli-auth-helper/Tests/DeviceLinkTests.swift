import Foundation
import CryptoKit

func runDeviceLinkTests() {
    let scope = InstallationScope(resource: "https://api.example.test", clientID: "riido-cli")
    let payload: [String: Any] = ["resource": scope.resource, "client_id": scope.clientID,
        "installation_id": "installation", "intent_id": "intent", "challenge_hash": String(repeating: "a", count: 64),
        "key_fingerprint": String(repeating: "b", count: 64), "expected_device_id": "device",
        "approved_principal_id": "approved-human", "expires_at": "2099-10-28T01:02:03.123456789Z"]
    let value = try! JSONSerialization.data(withJSONObject: payload)
    let request = try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": "installation-device-link-proof",
        "key": scope.storageKey, "interactive": true, "value": value.base64EncodedString()])
    var preflights = 0
    let handler = RequestHandler(preflight: { preflights += 1; throw HelperFailure.unavailable }, store: FakeStore(), confirm: {})
    check(handler.handle(request).error?.code == "UNAVAILABLE" && preflights == 1,
          "device-link proof reaches signature preflight before OS key access")
}

func runDeviceLinkSigningTests() {
    let now = Date(timeIntervalSince1970: 1793149000)
    let canonical = DeviceLinkProof(resource: "https://cli.fixture.invalid", clientID: "riido-cli",
        installationID: "installation", intentID: "intent", challengeHash: String(repeating: "a", count: 64),
        keyFingerprint: String(repeating: "b", count: 64), expectedDeviceID: "device",
        approvedPrincipalID: "approved-human", expiresAt: "2026-10-28T01:02:03.123Z")
    check(installationHex(installationDigest(canonical.framedMessage())) ==
        "979957be1deb56d1c5f4a68a6153ab690cc2c394d2551c3b0f1d7ed99db3a809", "device-link bytes match independent CP golden")
    let key = P256.Signing.PrivateKey()
    let fingerprint = installationHex(installationDigest(key.publicKey.x963Representation))
    var payload: [String: Any] = ["resource": canonical.resource, "client_id": "riido-cli", "installation_id": "installation",
        "intent_id": "intent", "challenge_hash": canonical.challengeHash, "key_fingerprint": fingerprint,
        "expected_device_id": "device", "approved_principal_id": "approved-human", "expires_at": "2026-10-28T01:02:03.123456789Z"]
    final class Keys: InstallationKeyStore {
        let key: P256.Signing.PrivateKey
        let now: Date
        var calls = 0
        var afterSign: (() -> Void)?
        init(_ key: P256.Signing.PrivateKey, _ now: Date) { self.key = key; self.now = now }
        func ensureKey(scope: InstallationScope) throws -> Data { check(false, "device link must not create keys"); throw HelperFailure.unavailable }
        func signProof(_ proof: InstallationProof) throws -> Data { check(false, "device link must not use OAuth domain"); throw HelperFailure.unavailable }
        func signDeviceLinkProof(_ proof: DeviceLinkProof) throws -> Data {
            calls += 1
            let signature = try key.signature(for: proof.message(publicKey: key.publicKey.x963Representation, now: now)).derRepresentation
            afterSign?()
            return signature
        }
    }
    let keys = Keys(key, now)
    let credentials = FakeStore()
    let handler = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { now })
    func request(_ object: [String: Any], interactive: Bool = true, storageKey: String? = nil) -> Data {
        let value = try! JSONSerialization.data(withJSONObject: object)
        return try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": "installation-device-link-proof",
            "key": storageKey ?? canonical.scope.storageKey, "interactive": interactive, "value": value.base64EncodedString()])
    }
    let response = handler.handle(request(payload))
    check(response.ok && keys.calls == 1, "device-link typed signer called")
    let encoded = try! JSONSerialization.data(withJSONObject: payload)
    let proof = try! DeviceLinkProof.decode(encoded, now: now)
    let text = response.data?.signature ?? ""
    let raw = (8...72).compactMap { try? installationDecodeURL(text, count: $0) }.first ?? Data()
    let signature = try? P256.Signing.ECDSASignature(derRepresentation: raw)
    check(signature != nil && key.publicKey.isValidSignature(signature!, for: try! proof.message(publicKey: key.publicKey.x963Representation, now: now)), "device-link DER verifies")
    check(handler.handle(request(payload, interactive: false)).error?.code == "INTERACTION_REQUIRED", "device-link requires interactive verification")
    check(!handler.handle(request(payload, storageKey: String(repeating: "0", count: 64))).ok, "device-link storage scope mismatch refused")
    for (field, value) in [("extra", "arbitrary-sign"), ("client_id", "other"), ("resource", "http://localhost"),
                           ("intent_id", "../intent"), ("expected_device_id", " device"), ("approved_principal_id", ""),
                           ("challenge_hash", "AA"), ("expires_at", "2026-10-28T01:02:03.120Z"),
                           ("expires_at", "2026-10-28T01:02:03+00:00"), ("expires_at", "2026-02-30T01:02:03Z"),
                           ("expires_at", "2020-01-01T00:00:00Z")] {
        var malformed = payload; malformed[field] = value
        check(!handler.handle(request(malformed)).ok, "device-link rejects malformed " + field)
    }
    check(keys.calls == 1 && credentials.calls.isEmpty, "invalid proof never accesses signing key or credentials")
    var currentTime = now
    keys.afterSign = { currentTime = try! deviceLinkExpiry(proof.expiresAt) }
    let slowHandler = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { currentTime })
    check(!slowHandler.handle(request(payload)).ok, "OS authentication reaching exact expiry cannot return signature")
    keys.afterSign = nil
    payload["key_fingerprint"] = String(repeating: "a", count: 64)
    check(!handler.handle(request(payload)).ok, "device-link key mismatch refuses signing")
    let duplicate = String(data: encoded, encoding: .utf8)!.dropLast() + ",\"client_id\":\"riido-cli\"}"
    check((try? DeviceLinkProof.decode(Data(duplicate.utf8), now: now)) == nil, "duplicate proof field rejected")
    check((try? DeviceLinkProof.decode(encoded, now: try! deviceLinkExpiry(proof.expiresAt))) == nil, "exact expiry refused")
}
