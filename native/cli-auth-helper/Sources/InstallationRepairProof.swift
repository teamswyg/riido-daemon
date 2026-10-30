import Foundation

// A fixed consent-request purpose, never an arbitrary message signer.
struct InstallationRepairProof: Decodable {
    let resource: String
    let clientID: String
    let requestID: String
    let keyFingerprint: String
    let recoveryHash: String
    enum CodingKeys: String, CodingKey {
        case resource
        case clientID = "client_id", requestID = "request_id"
        case keyFingerprint = "key_fingerprint", recoveryHash = "recovery_hash"
    }
    var scope: InstallationScope { InstallationScope(resource: resource, clientID: clientID) }
    func validate() throws {
        try scope.validate()
        guard resource.hasPrefix("https://") else { throw HelperFailure.invalid }
        _ = try installationDecodeURL(requestID, count: 32)
        _ = try installationDecodeDigest(keyFingerprint)
        _ = try installationDecodeDigest(recoveryHash)
    }
    func message(publicKey: Data) throws -> Data {
        try validate()
        _ = try installationPublicKey(publicKey)
        guard installationHex(installationDigest(publicKey)) == keyFingerprint else { throw HelperFailure.invalid }
        return installationFrame([Data("riido-cli-installation-repair-request.v1".utf8),
            Data(resource.utf8), Data(clientID.utf8), Data(requestID.utf8),
            try installationDecodeDigest(keyFingerprint), try installationDecodeDigest(recoveryHash)])
    }
    static func decode(_ bytes: Data) throws -> Self {
        let object = try flatJSONObject(bytes, maximum: 8192)
        let keys: Set<String> = ["resource", "client_id", "request_id", "key_fingerprint", "recovery_hash"]
        guard Set(object.keys) == keys,
              let proof = try? JSONDecoder().decode(Self.self, from: bytes) else { throw HelperFailure.invalid }
        try proof.validate()
        return proof
    }
}

extension InstallationKeyStore {
    func signRepairProof(_ proof: InstallationRepairProof) throws -> Data { throw HelperFailure.unavailable }
}
