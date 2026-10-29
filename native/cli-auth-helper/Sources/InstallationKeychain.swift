import Foundation
import Security
import LocalAuthentication
import CryptoKit

struct ProtectedInstallationKeyStore: InstallationKeyStore {
    func ensureKey(scope: InstallationScope) throws -> Data {
        try scope.validate()
        let context = installationContext()
        defer { context.invalidate() }
        let registration = InstallationKeyRegistration(
            reserve: { try InstallationKeyLock().acquire(key: scope.storageKey) },
            load: { try find(scope: scope, context: context) },
            create: { try create(scope: scope, context: context) }, publicBytes: publicBytes)
        return try registration.ensure()
    }

    private func create(scope: InstallationScope, context: LAContext) throws -> SecKey {
        var privateAttributes: [String: Any] = [
            kSecAttrIsPermanent as String: true,
            kSecAttrIsExtractable as String: false,
            kSecAttrApplicationTag as String: tag(scope),
            kSecAttrAccessGroup as String: SigningPolicy.accessGroup,
            kSecAttrAccessControl as String: try protectedAccessControl()
        ]
        privateAttributes[kSecAttrCanSign as String] = true
        let attributes: [String: Any] = [
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecAttrKeySizeInBits as String: 256,
            kSecUseDataProtectionKeychain as String: true,
            kSecAttrSynchronizable as String: false,
            kSecUseAuthenticationContext as String: context,
            kSecPrivateKeyAttrs as String: privateAttributes
        ]
        var error: Unmanaged<CFError>?
        if let key = SecKeyCreateRandomKey(attributes as CFDictionary, &error) { return key }
        throw securityFailure(error?.takeRetainedValue())
    }

    func signProof(_ proof: InstallationProof) throws -> Data {
        try proof.validate(now: Date())
        let context = installationContext()
        defer { context.invalidate() }
        let matches = try find(scope: proof.scope, context: context)
        if matches.isEmpty { throw HelperFailure.notFound }
        guard matches.count == 1 else { throw HelperFailure.unavailable }
        let key = matches[0]
        let publicKey = try publicBytes(key)
        let message = try proof.message(publicKey: publicKey, now: Date())
        guard SecKeyIsAlgorithmSupported(key, .sign, .ecdsaSignatureMessageX962SHA256) else { throw HelperFailure.unavailable }
        var error: Unmanaged<CFError>?
        guard let signature = SecKeyCreateSignature(key, .ecdsaSignatureMessageX962SHA256, message as CFData, &error) as Data? else {
            throw securityFailure(error?.takeRetainedValue())
        }
        // Authentication can outlive the intent. Never return a stale proof.
        try proof.validate(now: Date())
        guard signature.count <= 72,
              let der = try? P256.Signing.ECDSASignature(derRepresentation: signature),
              try installationPublicKey(publicKey).isValidSignature(der, for: message) else { throw HelperFailure.unavailable }
        return signature
    }

    private func tag(_ scope: InstallationScope) -> Data {
        Data(("io.riido.cli.installation.v1." + scope.storageKey).utf8)
    }
    private func installationContext() -> LAContext {
        let context = LAContext()
        context.localizedReason = "Verify the Riido CLI installation for your sign-in."
        context.touchIDAuthenticationAllowableReuseDuration = 0
        return context
    }
    private func find(scope: InstallationScope, context: LAContext) throws -> [SecKey] {
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrKeyClass as String: kSecAttrKeyClassPrivate,
            kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
            kSecAttrApplicationTag as String: tag(scope),
            kSecAttrAccessGroup as String: SigningPolicy.accessGroup,
            kSecAttrSynchronizable as String: false,
            kSecUseDataProtectionKeychain as String: true,
            kSecUseAuthenticationContext as String: context,
            kSecReturnRef as String: true, kSecMatchLimit as String: kSecMatchLimitAll
        ]
        var result: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecItemNotFound { return [] }
        guard status == errSecSuccess else { throw keychainFailure(status) }
        guard let values = result as? [AnyObject], values.count <= 1,
              values.allSatisfy({ CFGetTypeID($0) == SecKeyGetTypeID() }) else { throw HelperFailure.unavailable }
        // Match-all and checked CF identity prevent arbitrary key selection.
        return values.map { $0 as! SecKey }
    }
    private func publicBytes(_ privateKey: SecKey) throws -> Data {
        guard let attributes = SecKeyCopyAttributes(privateKey) as? [String: Any],
              attributes[kSecAttrKeyClass as String] as? String == kSecAttrKeyClassPrivate as String,
              attributes[kSecAttrKeyType as String] as? String == kSecAttrKeyTypeECSECPrimeRandom as String,
              attributes[kSecAttrKeySizeInBits as String] as? Int == 256,
              attributes[kSecAttrIsExtractable as String] as? Bool == false,
              let key = SecKeyCopyPublicKey(privateKey) else { throw HelperFailure.unavailable }
        var error: Unmanaged<CFError>?
        // Only the public key has an external representation; the private key
        // is never exported, logged, returned or persisted outside Keychain.
        guard let bytes = SecKeyCopyExternalRepresentation(key, &error) as Data? else {
            throw securityFailure(error?.takeRetainedValue())
        }
        _ = try installationPublicKey(bytes)
        return bytes
    }
    private func securityFailure(_ error: CFError?) -> HelperFailure {
        guard let error, CFErrorGetDomain(error) as String == NSOSStatusErrorDomain,
              let code = OSStatus(exactly: CFErrorGetCode(error)) else { return .unavailable }
        return keychainFailure(code)
    }
}
