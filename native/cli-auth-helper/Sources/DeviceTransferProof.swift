import Foundation

struct DeviceTransferProof {
    let link: DeviceLinkProof
    let previousDeviceID: String
    let previousLinkRevision: String
    var scope: InstallationScope { link.scope }

    func validate(now: Date) throws {
        try link.validate(now: now)
        _ = try installationDecodeURL(link.intentID, count: 32)
        guard deviceLinkID(previousDeviceID), previousDeviceID != link.expectedDeviceID,
              let revision = UInt64(previousLinkRevision), revision > 0, revision < UInt64.max,
              String(revision) == previousLinkRevision else { throw HelperFailure.invalid }
    }

    func message(publicKey: Data, now: Date) throws -> Data {
        try validate(now: now)
        _ = try link.message(publicKey: publicKey, now: now)
        return framedMessage()
    }

    func framedMessage() -> Data {
        installationFrame(["riido-cli-device-transfer.v1", "transfer", link.resource, link.clientID,
            link.installationID, link.intentID, link.challengeHash, link.keyFingerprint, link.expectedDeviceID,
            link.approvedPrincipalID, link.expiresAt, previousDeviceID, previousLinkRevision].map { Data($0.utf8) })
    }

    static func decode(_ bytes: Data, now: Date) throws -> Self {
        var fields = try flatJSONObject(bytes, maximum: 8192)
        let keys: Set<String> = ["resource", "client_id", "installation_id", "intent_id", "challenge_hash",
            "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at", "previous_device_id", "previous_link_revision"]
        guard Set(fields.keys) == keys,
              let device = fields.removeValue(forKey: "previous_device_id") as? String,
              let revision = fields.removeValue(forKey: "previous_link_revision") as? String else { throw HelperFailure.invalid }
        let link = try DeviceLinkProof.decode(JSONSerialization.data(withJSONObject: fields), now: now)
        let proof = Self(link: link, previousDeviceID: device, previousLinkRevision: revision)
        try proof.validate(now: now)
        return proof
    }
}

extension InstallationKeyStore {
    func signDeviceTransferProof(_ proof: DeviceTransferProof) throws -> Data { throw HelperFailure.unavailable }
}
