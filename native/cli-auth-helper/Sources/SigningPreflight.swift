import Foundation
import Security
import Darwin

func verifySigningPreflight() throws {
    guard isatty(STDIN_FILENO) == 0, isatty(STDOUT_FILENO) == 0 else { throw HelperFailure.unavailable }
    let bundle = Bundle.main
    guard bundle.bundleIdentifier == SigningPolicy.identifier,
          bundle.bundleURL.pathExtension == "app",
          let executable = bundle.executableURL,
          executable.deletingLastPathComponent().lastPathComponent == "MacOS"
    else { throw HelperFailure.unavailable }
    var code: SecCode?
    var staticCode: SecStaticCode?
    var information: CFDictionary?
    guard SecCodeCopySelf([], &code) == errSecSuccess, let code,
          SecCodeCopyStaticCode(code, [], &staticCode) == errSecSuccess, let staticCode,
          SecStaticCodeCheckValidity(staticCode, SecCSFlags(rawValue: kSecCSStrictValidate), nil) == errSecSuccess,
          SecCodeCopySigningInformation(staticCode, SecCSFlags(rawValue: kSecCSSigningInformation), &information) == errSecSuccess,
          let info = information as? [String: Any],
          let flags = info[kSecCodeInfoFlags as String] as? NSNumber,
          let entitlements = info[kSecCodeInfoEntitlementsDict as String] as? [String: Any]
    else { throw HelperFailure.unavailable }
    try SigningPolicy.validate(
        identifier: info[kSecCodeInfoIdentifier as String] as? String,
        team: info[kSecCodeInfoTeamIdentifier as String] as? String,
        entitlements: entitlements,
        adHoc: flags.uint32Value & SecCodeSignatureFlags.adhoc.rawValue != 0,
        hardened: flags.uint32Value & SecCodeSignatureFlags.runtime.rawValue != 0,
        profilePresent: FileManager.default.fileExists(atPath:
            bundle.bundleURL.appendingPathComponent("Contents/embedded.provisionprofile").path)
    )
}
