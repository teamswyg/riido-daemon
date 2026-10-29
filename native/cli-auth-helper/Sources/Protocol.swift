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
    enum CodingKeys: String, CodingKey {
        case present, value, confirmed, algorithm, signature
        case publicKey = "public_key", keyFingerprint = "key_fingerprint"
    }
}

struct HelperResponse: Encodable {
    let version = 1
    let ok: Bool
    var data: HelperData? = nil
    var error: HelperFailure? = nil
}

protocol CredentialStore {
    func status(key: String, interactive: Bool) throws -> Bool
    func load(key: String, interactive: Bool) throws -> Data
    func save(key: String, value: Data, interactive: Bool) throws
    func delete(key: String, interactive: Bool) throws
}
