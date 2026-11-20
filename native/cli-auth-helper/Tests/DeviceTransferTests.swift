import Foundation
import CryptoKit

func runDeviceTransferTests() {
    final class Keys: InstallationKeyStore {
        let key = P256.Signing.PrivateKey()
        var calls = 0
        var afterSign: (() -> Void)?
        func ensureKey(scope: InstallationScope) throws -> Data { check(false, "transfer cannot create keys"); throw HelperFailure.unavailable }
        func signProof(_ proof: InstallationProof) throws -> Data { check(false, "transfer cannot use OAuth domain"); throw HelperFailure.unavailable }
        func signDeviceTransferProof(_ proof: DeviceTransferProof) throws -> Data {
            calls += 1
            let signature = try key.signature(for: proof.message(publicKey: key.publicKey.x963Representation, now: bootstrapNow)).derRepresentation
            afterSign?()
            return signature
        }
    }
    let keys = Keys(), credentials = FakeStore()
    var payload = bootstrapPayload(fingerprint: installationHex(installationDigest(keys.key.publicKey.x963Representation)))
    for field in ["phase", "transition_id", "previous_boot_digest"] { payload.removeValue(forKey: field) }
    payload["previous_device_id"] = "source-device"; payload["previous_link_revision"] = "2"
    func request(_ value: [String: Any], interactive: Bool = true) -> Data {
        let scope = InstallationScope(resource: "https://cli.fixture.invalid", clientID: "riido-cli")
        return try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": "installation-device-transfer-proof",
            "key": scope.storageKey, "interactive": interactive, "value": JSONSerialization.data(withJSONObject: value).base64EncodedString()])
    }
    let handler = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { bootstrapNow })
    let response = handler.handle(request(payload))
    check(response.ok && keys.calls == 1, "transfer signed with existing key")
    let proof = try! DeviceTransferProof.decode(JSONSerialization.data(withJSONObject: payload), now: bootstrapNow)
    let raw = (8...72).compactMap { try? installationDecodeURL(response.data?.signature ?? "", count: $0) }.first ?? Data()
    let signature = try? P256.Signing.ECDSASignature(derRepresentation: raw)
    check(signature != nil && keys.key.publicKey.isValidSignature(signature!, for: proof.framedMessage()), "transfer DER binds exact source")
    check(signature != nil && !keys.key.publicKey.isValidSignature(signature!, for: proof.link.framedMessage()), "transfer proof cannot authorize initial link")
    for revision in ["", "0", "01", "+1", "-1", "18446744073709551615", "18446744073709551616"] {
        var invalid = payload; invalid["previous_link_revision"] = revision
        check(!handler.handle(request(invalid)).ok, "invalid source revision rejected")
    }
    for device in ["", "device", "../device", " source", "device\n"] {
        var invalid = payload; invalid["previous_device_id"] = device
        check(!handler.handle(request(invalid)).ok, "invalid or same source rejected")
    }
    var extra = payload; extra["signature_domain"] = "caller-selected"
    check(!handler.handle(request(extra)).ok, "transfer cannot choose signing domain")
    check(handler.handle(request(payload, interactive: false)).error?.code == "INTERACTION_REQUIRED", "transfer requires OS verification")
    check(keys.calls == 1 && credentials.calls.isEmpty, "invalid transfer never changes keys or credentials")
    var clock = bootstrapNow
    keys.afterSign = { clock = try! deviceLinkExpiry(proof.link.expiresAt) }
    let delayed = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { clock })
    check(!delayed.handle(request(payload)).ok, "transfer expiry rechecked after OS interaction")
    let encoded = try! JSONSerialization.data(withJSONObject: payload)
    check((try? DeviceLinkProof.decode(encoded, now: bootstrapNow)) == nil, "initial link decoder rejects transfer fields")
    let duplicate = String(data: encoded, encoding: .utf8)!.dropLast() + ",\"previous_link_revision\":\"3\"}"
    check((try? DeviceTransferProof.decode(Data(duplicate.utf8), now: bootstrapNow)) == nil, "duplicate source revision denied")
    payload["key_fingerprint"] = String(repeating: "b", count: 64)
    let golden = try! DeviceTransferProof.decode(JSONSerialization.data(withJSONObject: payload), now: bootstrapNow)
    check(installationHex(installationDigest(golden.framedMessage())) ==
        "f1dc69e70077293465bd980931f9e52eb55796fbf6793b537cc53a0180cf441c", "transfer matches current CP and Go frame")
}
