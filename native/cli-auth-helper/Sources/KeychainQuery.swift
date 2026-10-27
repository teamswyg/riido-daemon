import Foundation
import Security
import LocalAuthentication

func keychainQuery(key: String, interactive: Bool) -> [String: Any] {
    let context = LAContext()
    context.localizedReason = "Allow Riido CLI to access your Riido account credential."
    context.interactionNotAllowed = !interactive
    context.touchIDAuthenticationAllowableReuseDuration = 0
    return [kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: "io.riido.cli.userauth.v1",
            kSecAttrAccount as String: key,
            kSecAttrAccessGroup as String: SigningPolicy.accessGroup,
            kSecAttrSynchronizable as String: false,
            kSecUseDataProtectionKeychain as String: true,
            kSecUseAuthenticationContext as String: context]
}

func protectedAccessControl() throws -> SecAccessControl {
    var error: Unmanaged<CFError>?
    guard let control = SecAccessControlCreateWithFlags(nil,
        kSecAttrAccessibleWhenUnlockedThisDeviceOnly, .userPresence, &error)
    else {
        if let error { _ = error.takeRetainedValue() }
        throw HelperFailure.unavailable
    }
    return control
}

func keychainFailure(_ status: OSStatus) -> HelperFailure {
    switch status {
    case errSecItemNotFound: return .notFound
    case errSecInteractionNotAllowed, errSecAuthFailed: return .interactionRequired
    case errSecUserCanceled: return .cancelled
    default: return .unavailable
    }
}
