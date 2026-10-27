import Foundation

struct SigningPolicy {
    static let identifier = "io.riido.cli-auth-helper"
    static let team = "3PVV8DQPL6"
    static let accessGroup = team + "." + identifier

    static func validate(identifier: String?, team: String?, entitlements: [String: Any],
                         adHoc: Bool, hardened: Bool, profilePresent: Bool) throws {
        let applicationID = entitlements["com.apple.application-identifier"] as? String
            ?? entitlements["application-identifier"] as? String
        let groups = entitlements["keychain-access-groups"] as? [String]
        let unsafe = ["get-task-allow", "com.apple.security.get-task-allow",
                      "com.apple.security.cs.disable-library-validation",
                      "com.apple.security.cs.allow-unsigned-executable-memory",
                      "com.apple.security.cs.allow-dyld-environment-variables",
                      "com.apple.security.cs.disable-executable-page-protection",
                      "com.apple.security.cs.allow-jit"]
        guard identifier == Self.identifier, team == Self.team, !adHoc, hardened,
              profilePresent, applicationID == accessGroup,
              groups == [accessGroup],
              !unsafe.contains(where: { entitlements[$0] as? Bool == true })
        else { throw HelperFailure.unavailable }
    }
}
