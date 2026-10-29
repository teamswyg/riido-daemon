import Foundation
import CryptoKit

let installationAlgorithm = "p256-sha256-x962-der-v1"

func installationBase64URL(_ bytes: Data) -> String {
    bytes.base64EncodedString().replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_").replacingOccurrences(of: "=", with: "")
}

func installationDecodeURL(_ value: String, count: Int) throws -> Data {
    let padded = value.replacingOccurrences(of: "-", with: "+").replacingOccurrences(of: "_", with: "/")
        + String(repeating: "=", count: (4 - value.utf8.count % 4) % 4)
    guard let bytes = Data(base64Encoded: padded), bytes.count == count,
          installationBase64URL(bytes) == value else { throw HelperFailure.invalid }
    return bytes
}

func installationDigest(_ bytes: Data) -> Data { Data(SHA256.hash(data: bytes)) }
func installationHex(_ bytes: Data) -> String { bytes.map { String(format: "%02x", $0) }.joined() }

func installationDecodeDigest(_ value: String) throws -> Data {
    let bytes = Array(value.utf8)
    guard bytes.count == 64 else { throw HelperFailure.invalid }
    func nibble(_ byte: UInt8) throws -> UInt8 {
        if (48...57).contains(byte) { return byte - 48 }
        if (97...102).contains(byte) { return byte - 87 }
        throw HelperFailure.invalid
    }
    var result = Data()
    for offset in stride(from: 0, to: bytes.count, by: 2) {
        result.append(try nibble(bytes[offset]) * 16 + nibble(bytes[offset + 1]))
    }
    return result
}

func installationFrame(_ fields: [Data]) -> Data {
    var message = Data()
    for field in fields {
        var length = UInt32(field.count).bigEndian
        withUnsafeBytes(of: &length) { message.append(contentsOf: $0) }
        message.append(field)
    }
    return message
}

struct InstallationScope: Decodable {
    let resource: String
    let clientID: String
    enum CodingKeys: String, CodingKey { case resource; case clientID = "client_id" }

    func validate() throws {
        guard clientID == "riido-cli", resource.utf8.count <= 2048,
              let url = URLComponents(string: resource), let host = url.host, !host.isEmpty,
              url.user == nil, url.password == nil, url.query == nil, url.fragment == nil,
              url.path.isEmpty, url.percentEncodedPath.isEmpty,
              url.url?.absoluteString == resource,
              url.port == nil || (1...65535).contains(url.port!),
              url.scheme == "https" || url.scheme == "http" && ["localhost", "127.0.0.1", "[::1]"].contains(host)
        else { throw HelperFailure.invalid }
    }
    var storageKey: String {
        installationHex(installationDigest(installationFrame([
            Data("riido-cli-installation-key.v1".utf8), Data(resource.utf8), Data(clientID.utf8)
        ])))
    }
    static func decode(_ bytes: Data) throws -> Self {
        let object = try flatJSONObject(bytes, maximum: 8192)
        guard Set(object.keys) == ["resource", "client_id"],
              let scope = try? JSONDecoder().decode(Self.self, from: bytes) else { throw HelperFailure.invalid }
        try scope.validate()
        return scope
    }
}

struct InstallationProof: Decodable {
    let resource: String
    let clientID: String
    let intent: String
    let challengeHash: String
    let keyFingerprint: String
    let code: String
    let redirectURI: String
    let codeChallenge: String
    let state: String
    let expiresAt: Int64
    enum CodingKeys: String, CodingKey {
        case resource, code, state
        case clientID = "client_id", intent = "installation_intent", challengeHash = "challenge_hash"
        case keyFingerprint = "key_fingerprint", redirectURI = "redirect_uri", codeChallenge = "code_challenge"
        case expiresAt = "expires_at"
    }
    var scope: InstallationScope { InstallationScope(resource: resource, clientID: clientID) }
    func validate(now: Date) throws {
        try scope.validate()
        _ = try installationDecodeURL(intent, count: 32)
        _ = try installationDecodeURL(state, count: 32)
        _ = try installationDecodeURL(codeChallenge, count: 32)
        _ = try installationDecodeDigest(challengeHash)
        _ = try installationDecodeDigest(keyFingerprint)
        guard !code.isEmpty, code.utf8.count <= 256, redirectURI.utf8.count <= 2048,
              let redirect = URLComponents(string: redirectURI),
              redirect.scheme == "http", ["127.0.0.1", "[::1]"].contains(redirect.host ?? ""),
              let port = redirect.port, (1...65535).contains(port),
              redirect.user == nil, redirect.password == nil, redirect.query == nil, redirect.fragment == nil,
              redirect.percentEncodedPath == "/callback", redirect.url?.absoluteString == redirectURI,
              expiresAt > 0, Double(expiresAt) > now.timeIntervalSince1970
        else { throw HelperFailure.invalid }
    }
    func message(publicKey: Data, now: Date) throws -> Data {
        try validate(now: now)
        _ = try installationPublicKey(publicKey)
        guard installationHex(installationDigest(publicKey)) == keyFingerprint else { throw HelperFailure.invalid }
        return installationFrame([
            Data("riido-cli-installation-proof.v1".utf8), Data("authorization_code".utf8),
            Data(resource.utf8), Data(clientID.utf8), Data(intent.utf8),
            try installationDecodeDigest(challengeHash), try installationDecodeDigest(keyFingerprint),
            installationDigest(Data(code.utf8)), Data(redirectURI.utf8), Data(codeChallenge.utf8),
            Data("S256".utf8), installationDigest(Data(state.utf8)), Data("read".utf8), Data(String(expiresAt).utf8)
        ])
    }
    static func decode(_ bytes: Data, now: Date) throws -> Self {
        let object = try flatJSONObject(bytes, maximum: 8192)
        let keys: Set<String> = ["resource", "client_id", "installation_intent", "challenge_hash", "key_fingerprint",
                                 "code", "redirect_uri", "code_challenge", "state", "expires_at"]
        guard Set(object.keys) == keys, let proof = try? JSONDecoder().decode(Self.self, from: bytes) else { throw HelperFailure.invalid }
        try proof.validate(now: now)
        return proof
    }
}

func installationPublicKey(_ value: Data) throws -> P256.Signing.PublicKey {
    guard value.count == 65, value.first == 4,
          let key = try? P256.Signing.PublicKey(x963Representation: value) else { throw HelperFailure.unavailable }
    return key
}

protocol InstallationKeyStore {
    func ensureKey(scope: InstallationScope) throws -> Data
    func existingKey(scope: InstallationScope) throws -> Data
    func signProof(_ proof: InstallationProof) throws -> Data
    func signDeviceLinkProof(_ proof: DeviceLinkProof) throws -> Data
}

struct UnavailableInstallationKeyStore: InstallationKeyStore {
    func ensureKey(scope: InstallationScope) throws -> Data { throw HelperFailure.unavailable }
    func signProof(_ proof: InstallationProof) throws -> Data { throw HelperFailure.unavailable }
}
