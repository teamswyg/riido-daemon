import Foundation

struct HelperRequest: Decodable {
    let version: Int
    let operation: String
    let key: String
    let interactive: Bool
    let value: String?
}

struct HelperFailure: Error, Encodable {
    let code: String
    let message: String
    static let invalid = Self(code: "INVALID_REQUEST", message: "Invalid helper request.")
    static let unavailable = Self(code: "UNAVAILABLE", message: "Protected credential storage is unavailable.")
    static let interactionRequired = Self(code: "INTERACTION_REQUIRED", message: "User verification is required.")
    static let cancelled = Self(code: "CANCELLED", message: "User verification was cancelled.")
    static let timedOut = Self(code: "TIMED_OUT", message: "User verification timed out.")
    static let notFound = Self(code: "NOT_FOUND", message: "Credential not found.")
}

struct HelperData: Encodable {
    var present: Bool? = nil
    var value: String? = nil
    var confirmed: Bool? = nil
    var algorithm: String? = nil
    var publicKey: String? = nil
    var keyFingerprint: String? = nil
    var signature: String? = nil
    var bootDigest: String? = nil
    enum CodingKeys: String, CodingKey {
        case present, value, confirmed, algorithm, signature
        case publicKey = "public_key", keyFingerprint = "key_fingerprint"
        case bootDigest = "boot_digest"
    }
}

struct HelperResponse: Encodable {
    let version = 1
    let ok: Bool
    var data: HelperData? = nil
    var error: HelperFailure? = nil
}

// RIID-6014: the access grant item has no user-presence requirement, so a
// command whose access token is still valid reads it without OS confirmation.
// Refresh tokens and installation keys stay in the protected items.
protocol AccessCacheStore {
    func load(key: String) throws -> Data
    func save(key: String, value: Data) throws
    func delete(key: String) throws
}

struct UnavailableAccessCacheStore: AccessCacheStore {
    func load(key: String) throws -> Data { throw HelperFailure.unavailable }
    func save(key: String, value: Data) throws { throw HelperFailure.unavailable }
    func delete(key: String) throws { throw HelperFailure.unavailable }
}

protocol CredentialStore {
    func status(key: String, interactive: Bool) throws -> Bool
    func load(key: String, interactive: Bool) throws -> Data
    func save(key: String, value: Data, interactive: Bool) throws
    func delete(key: String, interactive: Bool) throws
}
