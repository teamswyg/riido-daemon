import Foundation
import LocalAuthentication
import Security

struct ProtectedKeychainStore: CredentialStore {
    func status(key: String, interactive: Bool) throws -> Bool {
        var query = keychainQuery(key: key, interactive: false)
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = false
        query[kSecReturnAttributes as String] = true
        return try protectedItemPresence(SecItemCopyMatching(query as CFDictionary, nil))
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

// RIID-6014: the access grant item. It is readable while the device is
// unlocked, never synchronized, limited to the helper's access group and kept
// apart from the user-presence protected credential by its service name. Its
// queries never allow UI.
struct ProtectedAccessCacheStore: AccessCacheStore {
    private func itemQuery(key: String) -> [String: Any] {
        let context = LAContext()
        context.interactionNotAllowed = true
        return [kSecClass as String: kSecClassGenericPassword,
                kSecAttrService as String: "io.riido.cli.userauth.access.v1",
                kSecAttrAccount as String: key,
                kSecAttrAccessGroup as String: SigningPolicy.accessGroup,
                kSecAttrSynchronizable as String: false,
                kSecUseDataProtectionKeychain as String: true,
                kSecUseAuthenticationContext as String: context]
    }
    func load(key: String) throws -> Data {
        var query = itemQuery(key: key)
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        query[kSecReturnData as String] = true
        var value: CFTypeRef?
        let result = SecItemCopyMatching(query as CFDictionary, &value)
        guard result == errSecSuccess else { throw keychainFailure(result) }
        guard let data = value as? Data else { throw HelperFailure.unavailable }
        return data
    }
    func save(key: String, value: Data) throws {
        let query = itemQuery(key: key)
        let attributes = [kSecValueData as String: value,
                          kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly] as [String: Any]
        let result = SecItemAdd(query.merging(attributes, uniquingKeysWith: { _, new in new }) as CFDictionary, nil)
        if result == errSecDuplicateItem {
            let updated = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
            guard updated == errSecSuccess else { throw keychainFailure(updated) }
        } else if result != errSecSuccess { throw keychainFailure(result) }
    }
    func delete(key: String) throws {
        let result = SecItemDelete(itemQuery(key: key) as CFDictionary)
        guard result == errSecSuccess || result == errSecItemNotFound else { throw keychainFailure(result) }
    }
}
