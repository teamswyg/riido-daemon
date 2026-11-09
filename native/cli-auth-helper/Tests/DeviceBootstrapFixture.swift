import Foundation
import CryptoKit

let bootstrapNow = Date(timeIntervalSince1970: 1793149000)
let bootstrapID = installationBase64URL(Data(repeating: 116, count: 32))
let bootstrapIntent = installationBase64URL(Data(repeating: 105, count: 32))

func bootstrapPayload(fingerprint: String) -> [String: Any] {
    ["resource": "https://cli.fixture.invalid", "client_id": "riido-cli",
     "installation_id": "installation", "intent_id": bootstrapIntent,
     "challenge_hash": String(repeating: "a", count: 64), "key_fingerprint": fingerprint,
     "expected_device_id": "device", "approved_principal_id": "approved-human",
     "expires_at": "2026-10-28T01:02:03.123Z", "phase": "prepare",
     "transition_id": bootstrapID, "previous_boot_digest": ""]
}

func bootstrapRequest(_ payload: [String: Any], interactive: Bool = true, storageKey: String? = nil) -> Data {
    let scope = InstallationScope(resource: "https://cli.fixture.invalid", clientID: "riido-cli")
    let value = try! JSONSerialization.data(withJSONObject: payload)
    return try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": "installation-device-bootstrap-proof",
        "key": storageKey ?? scope.storageKey, "interactive": interactive, "value": value.base64EncodedString()])
}

final class BootstrapTestKeys: InstallationKeyStore {
    let key = P256.Signing.PrivateKey()
    var calls = 0
    var bootDigest = String(repeating: "c", count: 64)
    var afterSign: (() -> Void)?
    func ensureKey(scope: InstallationScope) throws -> Data {
        check(false, "bootstrap must not create installation keys"); throw HelperFailure.unavailable
    }
    func signProof(_ proof: InstallationProof) throws -> Data {
        check(false, "bootstrap must not sign OAuth proofs"); throw HelperFailure.unavailable
    }
    func signDeviceBootstrapProof(_ proof: DeviceBootstrapProof) throws -> DeviceBootstrapSignature {
        calls += 1
        let message = try proof.message(publicKey: key.publicKey.x963Representation, bootDigest: bootDigest, now: bootstrapNow)
        let signature = try key.signature(for: message).derRepresentation
        afterSign?()
        return DeviceBootstrapSignature(bootDigest: bootDigest, signature: signature)
    }
}
