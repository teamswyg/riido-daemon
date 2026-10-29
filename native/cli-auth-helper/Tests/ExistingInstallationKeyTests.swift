import Foundation
import CryptoKit

func runExistingInstallationKeyTests() {
    final class Keys: InstallationKeyStore {
        let publicKey = P256.Signing.PrivateKey().publicKey.x963Representation
        var existingCalls = 0
        var ensureCalls = 0
        var failure: HelperFailure?
        func ensureKey(scope: InstallationScope) throws -> Data {
            ensureCalls += 1
            throw HelperFailure.unavailable
        }
        func existingKey(scope: InstallationScope) throws -> Data {
            existingCalls += 1
            if let failure { throw failure }
            return publicKey
        }
        func signProof(_ proof: InstallationProof) throws -> Data { throw HelperFailure.unavailable }
    }
    let scope = InstallationScope(resource: "https://api.example.test", clientID: "riido-cli")
    let credentials = FakeStore()
    let keys = Keys()
    var confirmations = 0
    let handler = RequestHandler(preflight: {}, store: credentials, confirm: { confirmations += 1 }, installationKeys: keys)
    func request(interactive: Bool = true, scopeKey: String? = nil, extraField: Bool = false) -> Data {
        var payload = ["resource": scope.resource, "client_id": scope.clientID]
        if extraField { payload["create_if_missing"] = "true" }
        let value = try! JSONSerialization.data(withJSONObject: payload)
        return try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": "installation-existing-key",
            "key": scopeKey ?? scope.storageKey, "interactive": interactive, "value": value.base64EncodedString()])
    }
    let found = handler.handle(request())
    check(found.ok && found.data?.publicKey == installationBase64URL(keys.publicKey), "existing lookup returns only existing public key")
    check(found.data?.keyFingerprint == installationHex(installationDigest(keys.publicKey)), "existing lookup binds public fingerprint")
    check(found.data?.value == nil && found.data?.signature == nil, "existing key response has no credential or signature")
    check(confirmations == 1 && keys.existingCalls == 1 && keys.ensureCalls == 0, "existing lookup confirms user without key creation")
    keys.failure = .notFound
    check(handler.handle(request()).error?.code == "NOT_FOUND", "missing existing key remains missing")
    check(keys.ensureCalls == 0 && keys.existingCalls == 2, "missing key cannot fall back to ensure")
    keys.failure = .unavailable
    check(handler.handle(request()).error?.code == "UNAVAILABLE", "ambiguous key fails closed")
    let before = keys.existingCalls
    check(handler.handle(request(interactive: false)).error?.code == "INTERACTION_REQUIRED", "existing lookup requires interaction")
    check(handler.handle(request(scopeKey: String(repeating: "0", count: 64))).error?.code == "INVALID_REQUEST", "existing lookup rejects foreign scope")
    check(handler.handle(request(extraField: true)).error?.code == "INVALID_REQUEST", "existing lookup rejects creation flag")
    let denied = RequestHandler(preflight: { throw HelperFailure.unavailable }, store: credentials, confirm: {}, installationKeys: keys)
    check(denied.handle(request()).error?.code == "UNAVAILABLE", "existing lookup verifies helper before key")
    let cancelled = RequestHandler(preflight: {}, store: credentials, confirm: { throw HelperFailure.cancelled }, installationKeys: keys)
    check(cancelled.handle(request()).error?.code == "CANCELLED", "existing lookup cancellation is preserved")
    check(keys.existingCalls == before && keys.ensureCalls == 0 && credentials.calls.isEmpty, "rejected lookup has no key or credential side effects")
}
