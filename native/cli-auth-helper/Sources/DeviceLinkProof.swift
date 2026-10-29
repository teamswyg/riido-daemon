import Foundation

// A closed server-issued device-link intent. This is deliberately separate
// from OAuth authorization proof and exposes no arbitrary signing primitive.
struct DeviceLinkProof: Decodable {
    let resource: String
    let clientID: String
    let installationID: String
    let intentID: String
    let challengeHash: String
    let keyFingerprint: String
    let expectedDeviceID: String
    let approvedPrincipalID: String
    let expiresAt: String
    enum CodingKeys: String, CodingKey {
        case resource
        case clientID = "client_id", installationID = "installation_id", intentID = "intent_id"
        case challengeHash = "challenge_hash", keyFingerprint = "key_fingerprint"
        case expectedDeviceID = "expected_device_id", approvedPrincipalID = "approved_principal_id"
        case expiresAt = "expires_at"
    }
    var scope: InstallationScope { InstallationScope(resource: resource, clientID: clientID) }
    func validate(now: Date) throws {
        try scope.validate()
        guard resource.hasPrefix("https://"),
              [installationID, intentID, expectedDeviceID, approvedPrincipalID].allSatisfy(deviceLinkID),
              try deviceLinkExpiry(expiresAt) > now else { throw HelperFailure.invalid }
        _ = try installationDecodeDigest(challengeHash)
        _ = try installationDecodeDigest(keyFingerprint)
    }
    func message(publicKey: Data, now: Date) throws -> Data {
        try validate(now: now)
        _ = try installationPublicKey(publicKey)
        guard installationHex(installationDigest(publicKey)) == keyFingerprint else { throw HelperFailure.invalid }
        return framedMessage()
    }
    // Digests are framed lowercase hex text in the CP contract, unlike OAuth.
    func framedMessage() -> Data {
        installationFrame(["riido-cli-device-link.v1", "link", resource, clientID, installationID, intentID,
                           challengeHash, keyFingerprint, expectedDeviceID, approvedPrincipalID, expiresAt]
            .map { Data($0.utf8) })
    }
    static func decode(_ bytes: Data, now: Date) throws -> Self {
        let object = try flatJSONObject(bytes, maximum: 8192)
        let keys: Set<String> = ["resource", "client_id", "installation_id", "intent_id", "challenge_hash",
                                 "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at"]
        guard Set(object.keys) == keys, let proof = try? JSONDecoder().decode(Self.self, from: bytes) else { throw HelperFailure.invalid }
        try proof.validate(now: now)
        return proof
    }
}

private func deviceLinkID(_ value: String) -> Bool {
    !value.isEmpty && value.utf8.count <= 512 && value.trimmingCharacters(in: .whitespacesAndNewlines) == value
        && !value.contains(where: { "/\\\0\r\n\t".contains($0) })
}

// Preserve the original canonical nanosecond text in the signed message. Date
// formatters cannot round-trip all nine fractional digits without precision loss.
func deviceLinkExpiry(_ value: String) throws -> Date {
    let pattern = "^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]{0,8}[1-9])?Z$"
    guard value.utf8.count <= 30, value.range(of: pattern, options: .regularExpression) != nil else { throw HelperFailure.invalid }
    let base = String(value.prefix(19)) + "Z"
    let formatter = ISO8601DateFormatter()
    formatter.formatOptions = [.withInternetDateTime]
    guard let date = formatter.date(from: base), formatter.string(from: date) == base else { throw HelperFailure.invalid }
    if value.count == 20 { return date }
    let fraction = "0" + value.dropFirst(19).dropLast()
    guard let seconds = Double(fraction) else { throw HelperFailure.invalid }
    return date.addingTimeInterval(seconds)
}

extension InstallationKeyStore {
    // Older/fake stores do not silently treat another proof domain as OAuth.
    func signDeviceLinkProof(_ proof: DeviceLinkProof) throws -> Data { throw HelperFailure.unavailable }
}
