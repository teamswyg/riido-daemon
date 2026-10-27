import Foundation
import LocalAuthentication

protocol PresenceContext {
    func evaluate(_ completion: @escaping (Bool, Error?) -> Void)
    func invalidate()
}

final class NativePresenceContext: PresenceContext {
    private let context = LAContext()
    init() { context.touchIDAuthenticationAllowableReuseDuration = 0 }
    func evaluate(_ completion: @escaping (Bool, Error?) -> Void) {
        context.evaluatePolicy(.deviceOwnerAuthentication,
                               localizedReason: "Riido CLI에 로그인하려면 사용자 확인이 필요합니다.", reply: completion)
    }
    func invalidate() { context.invalidate() }
}

struct UserPresenceVerifier {
    var timeout: TimeInterval = 120
    var makeContext: () -> PresenceContext = { NativePresenceContext() }

    func confirm() throws {
        let context = makeContext()
        defer { context.invalidate() }
        let pending = PendingPresence()
        context.evaluate { success, error in pending.resolve(success: success, error: error) }
        // LA replies run on a private framework queue. Do not reuse its context
        // or invoke another policy check from inside that callback.
        guard pending.ready.wait(timeout: .now() + max(0, min(timeout, 120))) == .success else {
            pending.discard()
            throw HelperFailure.timedOut
        }
        try pending.result()
    }
}

private final class PendingPresence {
    let ready = DispatchSemaphore(value: 0)
    private let lock = NSLock()
    private var pending = true
    private var outcome: Result<Void, HelperFailure>?

    func resolve(success: Bool, error: Error?) {
        lock.lock()
        defer { lock.unlock() }
        guard pending else { return }
        pending = false
        if success && error == nil { outcome = .success(()) }
        else { outcome = .failure(presenceFailure(error)) }
        ready.signal()
    }
    func discard() {
        lock.lock()
        defer { lock.unlock() }
        pending = false
        outcome = nil
    }
    func result() throws {
        lock.lock()
        let captured = outcome
        lock.unlock()
        guard let captured else { throw HelperFailure.unavailable }
        try captured.get()
    }
}

private func presenceFailure(_ error: Error?) -> HelperFailure {
    guard let failure = error as? LAError else { return .unavailable }
    switch failure.code {
    case .userCancel, .appCancel, .systemCancel: return .cancelled
    case .authenticationFailed: return .interactionRequired
    default: return .unavailable
    }
}
