import Foundation
import LocalAuthentication

if CommandLine.arguments.dropFirst().first == "--installation-lock-probe" { exit(installationLockProbe()) }

var failures = 0
func check(_ condition: @autoclosure () -> Bool, _ message: String) {
    if !condition() { failures += 1; fputs("FAIL: \(message)\n", stderr) }
}
final class FakeStore: CredentialStore {
    var calls = [String]()
    var value: Data? = Data("fixture-only".utf8)
    func status(key: String, interactive: Bool) throws -> Bool {
        calls.append("status"); return value != nil
    }
    func load(key: String, interactive: Bool) throws -> Data {
        calls.append("load")
        guard let value else { throw HelperFailure.notFound }; return value
    }
    func save(key: String, value: Data, interactive: Bool) throws {
        calls.append("save"); self.value = value
    }
    func delete(key: String, interactive: Bool) throws { calls.append("delete"); value = nil }
}
let key = "fixture_account_key_5900"
func request(_ operation: String, value: String? = nil, interactive: Bool = false) -> Data {
    var json: [String: Any] = ["version": 1, "operation": operation, "key": key, "interactive": interactive]
    if let value { json["value"] = value }
    return try! JSONSerialization.data(withJSONObject: json)
}
let store = FakeStore()
var confirmCalls = 0
let handler = RequestHandler(preflight: {}, store: store, confirm: { confirmCalls += 1 })
check(handler.handle(request("status")).data?.present == true, "status sees presence")
check(store.calls == ["status"], "status never loads bytes")
check(handler.handle(request("status")).data?.value == nil, "status omits secret")
let loaded = handler.handle(request("load"))
check(loaded.ok && Data(base64Encoded: loaded.data?.value ?? "") == store.value, "load private value")
let marker = "not-a-token"
check(handler.handle(request("save", value: Data(marker.utf8).base64EncodedString())).ok, "save")
check(store.value == Data(marker.utf8), "save exact bytes")
check(handler.handle(request("delete")).ok && store.value == nil, "delete")
check(handler.handle(request("load")).error?.code == "NOT_FOUND", "missing mapping")
let beforeInvalid = store.calls.count
for invalid in [Data(), Data("{}".utf8), request("unknown"), request("save"), request("save", value: "%%%%"), request("load", value: "eA=="), Data(repeating: 32, count: 65537)] {
    check(handler.handle(invalid).error?.code == "INVALID_REQUEST", "invalid request")
}
check(store.calls.count == beforeInvalid, "invalid request never touches store")
let blocked = RequestHandler(preflight: { throw HelperFailure.unavailable }, store: store, confirm: { confirmCalls += 1 })
let denied = blocked.handle(request("load"))
check(denied.error?.code == "UNAVAILABLE", "failed preflight")
check(store.calls.count == beforeInvalid, "failed preflight never touches store")
let encoded = String(data: try! JSONEncoder().encode(denied), encoding: .utf8)!
check(!encoded.contains(marker), "error never serializes private value")
check(handler.handle(request("save", value: Data(repeating: 7, count: 32769).base64EncodedString())).error?.code == "INVALID_REQUEST", "value bound")
var validEntitlements: [String: Any] = [
    "com.apple.application-identifier": SigningPolicy.accessGroup,
    "keychain-access-groups": [SigningPolicy.accessGroup]
]
func acceptsSignature(team: String = SigningPolicy.team, identifier: String = SigningPolicy.identifier,
                      adHoc: Bool = false, hardened: Bool = true, profile: Bool = true,
                      entitlements: [String: Any] = validEntitlements) -> Bool {
    do {
        try SigningPolicy.validate(identifier: identifier, team: team, entitlements: entitlements,
                                   adHoc: adHoc, hardened: hardened, profilePresent: profile)
        return true
    } catch { return false }
}
check(acceptsSignature(), "expected signing metadata")
check(!acceptsSignature(team: "OTHERTEAM"), "reject other developer team")
check(!acceptsSignature(identifier: "io.attacker.helper"), "reject other binary")
check(!acceptsSignature(adHoc: true), "reject ad hoc")
check(!acceptsSignature(hardened: false), "require hardened runtime")
check(!acceptsSignature(profile: false), "require embedded profile")
for name in ["get-task-allow", "com.apple.security.get-task-allow",
             "com.apple.security.cs.disable-library-validation", "com.apple.security.cs.allow-unsigned-executable-memory",
             "com.apple.security.cs.allow-dyld-environment-variables",
             "com.apple.security.cs.disable-executable-page-protection", "com.apple.security.cs.allow-jit"] {
    var entitlements = validEntitlements; entitlements[name] = true
    check(!acceptsSignature(entitlements: entitlements), "reject unsafe entitlement")
}
var extraGroup = validEntitlements
extraGroup["keychain-access-groups"] = [SigningPolicy.accessGroup, "3PVV8DQPL6.other"]
check(!acceptsSignature(entitlements: extraGroup), "dedicated keychain access group")
for replacement in ["../fixture_account_5900", "short", String(repeating: "x", count: 129), "한글계정키는허용하지않음"] {
    var object = try! JSONSerialization.jsonObject(with: request("status")) as! [String: Any]
    object["key"] = replacement
    check(handler.handle(try! JSONSerialization.data(withJSONObject: object)).error?.code == "INVALID_REQUEST", "opaque bounded key")
}
var extra = try! JSONSerialization.jsonObject(with: request("status")) as! [String: Any]
extra["debug"] = true
check(handler.handle(try! JSONSerialization.data(withJSONObject: extra)).error?.code == "INVALID_REQUEST", "reject unknown fields")
for failure in [HelperFailure.interactionRequired, .cancelled, .notFound] {
    let refused = RequestHandler(preflight: { throw failure }, store: store, confirm: { confirmCalls += 1 }).handle(request("load"))
    check(refused.error?.code == failure.code && refused.data == nil, "sanitized refusal")
}
let beforeConfirm = store.calls.count
let confirmation = handler.handle(request("confirm", interactive: true))
check(confirmation.ok && confirmation.data?.confirmed == true, "explicit first-login confirmation is supported")
check(confirmation.data?.present == nil && confirmation.data?.value == nil, "confirmation contains no credential evidence")
check(confirmCalls == 1, "interactive confirmation invokes OS verifier exactly once")
check(handler.handle(request("confirm")).error?.code == "INTERACTION_REQUIRED", "no-input confirmation refuses")
check(confirmCalls == 1, "no-input does not invoke OS verifier")
check(blocked.handle(request("confirm", interactive: true)).error?.code == "UNAVAILABLE", "signature preflight before confirmation")
check(confirmCalls == 1, "signature failure does not invoke OS verifier")
check(handler.handle(request("confirm", value: "eA==", interactive: true)).error?.code == "INVALID_REQUEST", "confirm rejects credential input")
for failure in [HelperFailure.cancelled, .interactionRequired, .unavailable] {
    let refusing = RequestHandler(preflight: {}, store: store, confirm: { throw failure })
    check(refusing.handle(request("confirm", interactive: true)).error?.code == failure.code, "confirmation refusal is sanitized")
}
let confirmJSON = String(data: try! JSONEncoder().encode(confirmation), encoding: .utf8)!
check(!confirmJSON.contains("present") && !confirmJSON.contains("value"), "confirm JSON has no storage fields")
check(store.calls.count == beforeConfirm, "confirmation never accesses keychain storage")
final class FakePresenceContext: PresenceContext {
    var calls = 0
    var invalidations = 0
    var callback: ((Bool, Error?) -> Void)?
    var response: (Bool, Error?)? = (true, nil)
    func evaluate(_ completion: @escaping (Bool, Error?) -> Void) {
        calls += 1
        callback = completion
        if let response { completion(response.0, response.1) }
    }
    func invalidate() { invalidations += 1 }
}
var contexts = [FakePresenceContext]()
let verifier = UserPresenceVerifier(makeContext: {
    let context = FakePresenceContext()
    contexts.append(context)
    return context
})
for _ in 0..<2 { do { try verifier.confirm() } catch { check(false, "fresh confirmation succeeds") } }
check(contexts.count == 2 && contexts.allSatisfy { $0.calls == 1 && $0.invalidations == 1 }, "fresh context per confirmation and no cache")
for (code, expected) in [(LAError.Code.userCancel, "CANCELLED"), (.appCancel, "CANCELLED"), (.systemCancel, "CANCELLED"), (.authenticationFailed, "INTERACTION_REQUIRED"), (.invalidContext, "UNAVAILABLE")] {
    let context = FakePresenceContext()
    context.response = (false, LAError(code))
    do { try UserPresenceVerifier(makeContext: { context }).confirm(); check(false, "OS refusal must throw") }
    catch let failure as HelperFailure { check(failure.code == expected, "OS refusal mapping") }
    catch { check(false, "OS error escaped") }
    check(context.invalidations == 1, "refusal invalidates context")
}
for response: (Bool, Error?) in [(false, nil), (true, LAError(.userCancel)), (false, NSError(domain: "private-native-error", code: 1))] {
    let context = FakePresenceContext()
    context.response = response
    do { try UserPresenceVerifier(makeContext: { context }).confirm(); check(false, "ambiguous response cannot confirm") }
    catch is HelperFailure {} catch { check(false, "raw OS error escaped") }
    check(context.invalidations == 1, "invalid response invalidates context")
}
let timedOut = FakePresenceContext()
timedOut.response = nil
let started = Date()
do { try UserPresenceVerifier(timeout: 0.01, makeContext: { timedOut }).confirm(); check(false, "missing reply times out") }
catch let failure as HelperFailure { check(failure.code == "TIMED_OUT", "timeout refuses separately from cancellation") }
catch { check(false, "timeout error escaped") }
check(Date().timeIntervalSince(started) < 2 && timedOut.invalidations == 1, "bounded wait invalidates context")
timedOut.callback?(true, nil)
check(timedOut.invalidations == 1, "late success cannot reopen confirmation")
let installationRequest = try! JSONSerialization.data(withJSONObject: [
    "version": 1, "operation": "installation-key", "key": "08eba32e93e89b49cbed7c040f8de7dba724391c4015de3baaa78d9dad03957f", "interactive": true,
    "value": Data("{\"resource\":\"https://api.example.test\",\"client_id\":\"riido-cli\"}".utf8).base64EncodedString()
])
check(blocked.handle(installationRequest).error?.code == "UNAVAILABLE", "typed installation key reaches signature preflight before key access")
runInstallationTests()
runDeviceLinkTests()
runDeviceLinkSigningTests()
runInstallationRegistrationTests()
runInstallationLockTests()
if failures > 0 { exit(1) }
print("PASS: helper protocol fixtures (no real keychain access)")
