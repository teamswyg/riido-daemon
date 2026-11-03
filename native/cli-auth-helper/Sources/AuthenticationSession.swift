import Foundation
import LocalAuthentication

// One helper process serves one CLI command in session mode. The first
// protected operation asks the OS for user confirmation once with this shared
// context; later keychain and Secure Enclave operations in the same command
// reuse that evaluated context instead of prompting again. The context dies
// with the helper process (stdin EOF, request limit or lifetime limit).
// Set only by the session entry point; per-request mode leaves it nil.
var currentAuthenticationSession: AuthenticationSession?

final class AuthenticationSession {
    let context: LAContext
    private let verifier: UserPresenceVerifier
    private var verified = false

    init(context: LAContext = AuthenticationSession.makeContext(), verifier: UserPresenceVerifier? = nil) {
        self.context = context
        self.verifier = verifier ?? UserPresenceVerifier(makeContext: { SharedPresenceContext(context: context) })
    }

    static func makeContext() -> LAContext {
        let context = LAContext()
        context.localizedReason = "Riido CLI에서 계정 정보를 사용하려면 사용자 확인이 필요합니다."
        context.touchIDAuthenticationAllowableReuseDuration = 0
        return context
    }

    func ensureVerified() throws {
        if verified { return }
        try verifier.confirm()
        verified = true
    }

    func invalidate() { context.invalidate() }
}

// The shared context outlives a single confirmation, so it is not invalidated
// after each check; AuthenticationSession owns its lifetime.
final class SharedPresenceContext: PresenceContext {
    private let context: LAContext
    init(context: LAContext) { self.context = context }
    func evaluate(_ completion: @escaping (Bool, Error?) -> Void) {
        context.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: context.localizedReason, reply: completion)
    }
    func invalidate() {}
}
