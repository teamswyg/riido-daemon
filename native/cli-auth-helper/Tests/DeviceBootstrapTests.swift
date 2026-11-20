import Foundation
import CryptoKit

func runDeviceBootstrapTests() {
    let keys = BootstrapTestKeys(), credentials = FakeStore()
    let fingerprint = installationHex(installationDigest(keys.key.publicKey.x963Representation))
    let payload = bootstrapPayload(fingerprint: fingerprint)
    let handler = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { bootstrapNow })
    let response = handler.handle(bootstrapRequest(payload))
    check(response.ok && keys.calls == 1 && response.data?.bootDigest == keys.bootDigest, "bootstrap signs fixed prepare intent")
    let proof = try! DeviceBootstrapProof.decode(JSONSerialization.data(withJSONObject: payload), now: bootstrapNow)
    let raw = (8...72).compactMap { try? installationDecodeURL(response.data?.signature ?? "", count: $0) }.first ?? Data()
    let signature = try? P256.Signing.ECDSASignature(derRepresentation: raw)
    check(signature != nil && keys.key.publicKey.isValidSignature(signature!, for:
        try! proof.message(publicKey: keys.key.publicKey.x963Representation, bootDigest: keys.bootDigest, now: bootstrapNow)), "bootstrap DER signature binds boot witness")
    check(handler.handle(bootstrapRequest(payload, interactive: false)).error?.code == "INTERACTION_REQUIRED", "bootstrap requires OS interaction")
    check(!handler.handle(bootstrapRequest(payload, storageKey: String(repeating: "0", count: 64))).ok, "bootstrap storage mismatch denied")
    for (field, value) in [("phase", "unknown"), ("transition_id", "legacy"), ("intent_id", "legacy"),
        ("previous_boot_digest", keys.bootDigest), ("boot_digest", keys.bootDigest), ("boot_id", "caller-controlled"),
        ("expires_at", "2020-01-01T00:00:00Z"), ("resource", "http://localhost"), ("key_fingerprint", "invalid")] {
        var invalid = payload; invalid[field] = value
        check(!handler.handle(bootstrapRequest(invalid)).ok, "bootstrap refuses malformed " + field)
    }
    check(keys.calls == 1 && credentials.calls.isEmpty, "invalid bootstrap does not touch signing key or credentials")
    var activate = payload; activate["phase"] = "activate"
    activate["previous_boot_digest"] = String(repeating: "a", count: 64)
    check(handler.handle(bootstrapRequest(activate)).ok, "different boot activation succeeds")
    activate["previous_boot_digest"] = keys.bootDigest
    check(!handler.handle(bootstrapRequest(activate)).ok, "same boot activation denied")
    var clock = bootstrapNow
    keys.afterSign = { clock = try! deviceLinkExpiry(proof.link.expiresAt) }
    let delayed = RequestHandler(preflight: {}, store: credentials, confirm: {}, installationKeys: keys, now: { clock })
    check(!delayed.handle(bootstrapRequest(payload)).ok, "OS delay cannot extend bootstrap expiry")
    var preflights = 0
    let untrusted = RequestHandler(preflight: { preflights += 1; throw HelperFailure.unavailable }, store: credentials, confirm: {}, now: { bootstrapNow })
    check(untrusted.handle(bootstrapRequest(payload)).error?.code == "UNAVAILABLE" && preflights == 1, "bootstrap code trust before protected access")
    check(credentials.calls.isEmpty, "bootstrap preserves saved credentials")
    let encoded = String(data: try! JSONSerialization.data(withJSONObject: payload), encoding: .utf8)!
    check((try? DeviceBootstrapProof.decode(Data((String(encoded.dropLast()) + ",\"phase\":\"activate\"}").utf8), now: bootstrapNow)) == nil, "duplicate phase denied")
}

func runBootSessionTests() {
    // Read only: no keychain, user prompt, reboot, raw identifier output or storage.
    let current = try? currentBootSessionID()
    check(current != nil && current == (try? currentBootSessionID()), "kernel boot identity is stable within this boot")
    for bad in ["", "not-a-uuid", "00000000-0000-0000-0000-000000000000", " 11111111-1111-4111-8111-111111111111"] {
        check((try? canonicalBootSessionID(bad)) == nil, "invalid boot identity denied")
    }
    var payload = bootstrapPayload(fingerprint: String(repeating: "b", count: 64))
    let proof = try! DeviceBootstrapProof.decode(JSONSerialization.data(withJSONObject: payload), now: bootstrapNow)
    let first = try! scopedBootSessionDigest(proof, bootID: "11111111-1111-4111-8111-111111111111")
    check(installationHex(installationDigest(try! proof.framedMessage(bootDigest: String(repeating: "c", count: 64)))) ==
        "3cc36ae5f4cf0c3264b303853809ae2d0085eb466e3eedda70651ca17c06d009", "bootstrap matches independent Go frame")
    let second = try! scopedBootSessionDigest(proof, bootID: "22222222-2222-4222-8222-222222222222")
    check(first != second && first.count == 64, "boot witness changes across boot sessions")
    for (field, value) in [("resource", "https://other.fixture.invalid"), ("key_fingerprint", String(repeating: "c", count: 64))] {
        payload = bootstrapPayload(fingerprint: String(repeating: "b", count: 64)); payload[field] = value
        let changed = try! DeviceBootstrapProof.decode(JSONSerialization.data(withJSONObject: payload), now: bootstrapNow)
        check(first != (try! scopedBootSessionDigest(changed, bootID: "11111111-1111-4111-8111-111111111111")), "boot witness isolated by " + field)
    }
}
