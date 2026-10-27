import Foundation
import Security

struct ProtectedKeychainStore: CredentialStore {
    func status(key: String, interactive: Bool) throws -> Bool {
        var query = keychainQuery(key: key, interactive: false)
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = false
        query[kSecReturnAttributes as String] = true
        let result = SecItemCopyMatching(query as CFDictionary, nil)
        if result == errSecItemNotFound { return false }
        guard result == errSecSuccess else { throw keychainFailure(result) }
        return true
    }
    func load(key: String, interactive: Bool) throws -> Data {
        var query = keychainQuery(key: key, interactive: interactive)
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = true
        var value: CFTypeRef?
        let result = SecItemCopyMatching(query as CFDictionary, &value)
        guard result == errSecSuccess else { throw keychainFailure(result) }
        guard let data = value as? Data else { throw HelperFailure.unavailable }
        return data
    }
    func save(key: String, value: Data, interactive: Bool) throws {
        let query = keychainQuery(key: key, interactive: interactive)
        let attributes = [kSecValueData as String: value,
                          kSecAttrAccessControl as String: try protectedAccessControl()] as [String: Any]
        let addition = query.merging(attributes, uniquingKeysWith: { _, new in new })
        let result = SecItemAdd(addition as CFDictionary, nil)
        if result == errSecDuplicateItem {
            let updated = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
            guard updated == errSecSuccess else { throw keychainFailure(updated) }
        } else if result != errSecSuccess { throw keychainFailure(result) }
    }
    func delete(key: String, interactive: Bool) throws {
        let result = SecItemDelete(keychainQuery(key: key, interactive: interactive) as CFDictionary)
        guard result == errSecSuccess || result == errSecItemNotFound else { throw keychainFailure(result) }
    }
}
