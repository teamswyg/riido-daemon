import Foundation

struct DeviceBootstrapProof {
    let link: DeviceLinkProof
    let phase: String
    let transitionID: String
    let previousBootDigest: String
    var scope: InstallationScope { link.scope }

    func validate(now: Date) throws {
        try link.validate(now: now)
        _ = try installationDecodeURL(transitionID, count: 32)
        _ = try installationDecodeURL(link.intentID, count: 32)
        if phase == "prepare" {
            guard previousBootDigest.isEmpty else { throw HelperFailure.invalid }
        } else if phase == "activate" {
            _ = try installationDecodeDigest(previousBootDigest)
        } else { throw HelperFailure.invalid }
    }

    func message(publicKey: Data, bootDigest: String, now: Date) throws -> Data {
        try validate(now: now)
        _ = try link.message(publicKey: publicKey, now: now)
        return try framedMessage(bootDigest: bootDigest)
    }

    func framedMessage(bootDigest: String) throws -> Data {
        _ = try installationDecodeDigest(bootDigest)
        guard phase != "activate" || previousBootDigest != bootDigest else { throw HelperFailure.invalid }
        return installationFrame(["riido-cli-device-bootstrap.v1", phase, link.resource, link.clientID,
            link.installationID, transitionID, link.intentID, link.challengeHash, link.keyFingerprint,
            link.expectedDeviceID, link.approvedPrincipalID, link.expiresAt, previousBootDigest, bootDigest]
            .map { Data($0.utf8) })
    }

    static func decode(_ bytes: Data, now: Date) throws -> Self {
        var fields = try flatJSONObject(bytes, maximum: 8192)
        let keys: Set<String> = ["resource", "client_id", "installation_id", "intent_id", "challenge_hash",
            "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at",
            "phase", "transition_id", "previous_boot_digest"]
        guard Set(fields.keys) == keys,
              let phase = fields.removeValue(forKey: "phase") as? String,
              let transition = fields.removeValue(forKey: "transition_id") as? String,
              let previous = fields.removeValue(forKey: "previous_boot_digest") as? String else { throw HelperFailure.invalid }
        let linkBytes = try JSONSerialization.data(withJSONObject: fields)
        let proof = Self(link: try DeviceLinkProof.decode(linkBytes, now: now), phase: phase,
                         transitionID: transition, previousBootDigest: previous)
        try proof.validate(now: now)
        return proof
    }
}

struct DeviceBootstrapSignature {
    let bootDigest: String
    let signature: Data
}

extension InstallationKeyStore {
    func signDeviceBootstrapProof(_ proof: DeviceBootstrapProof) throws -> DeviceBootstrapSignature { throw HelperFailure.unavailable }
}
