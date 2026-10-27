import Foundation

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
func request(_ operation: String, value: String? = nil) -> Data {
    var json: [String: Any] = ["version": 1, "operation": operation, "key": key, "interactive": false]
    if let value { json["value"] = value }
    return try! JSONSerialization.data(withJSONObject: json)
}
let store = FakeStore()
let handler = RequestHandler(preflight: {}, store: store)
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
let blocked = RequestHandler(preflight: { throw HelperFailure.unavailable }, store: store)
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
    let refused = RequestHandler(preflight: { throw failure }, store: store).handle(request("load"))
    check(refused.error?.code == failure.code && refused.data == nil, "sanitized refusal")
}
if failures > 0 { exit(1) }
print("PASS: helper protocol fixtures (no real keychain access)")
