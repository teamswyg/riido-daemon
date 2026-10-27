import Foundation
import CryptoKit

func runInstallationTests() {
    struct Golden: Decodable {
        let IntentID, Resource, ClientID, RedirectURI, Scope, State, CodeChallenge: String
        let ChallengeHash, PublicKey, KeyFingerprint, Code, MessageBase64: String
        let ExpiresAt: Int64
    }
    let vectorURL = URL(fileURLWithPath: CommandLine.arguments[1])
    let golden = try! JSONDecoder().decode(Golden.self, from: Data(contentsOf: vectorURL))
    let now = Date(timeIntervalSince1970: Double(golden.ExpiresAt - 300))
    let publicKey = try! installationDecodeURL(golden.PublicKey, count: 65)
    let scope = InstallationScope(resource: golden.Resource, clientID: golden.ClientID)
    let proof = InstallationProof(resource: golden.Resource, clientID: golden.ClientID, intent: golden.IntentID,
                                  challengeHash: golden.ChallengeHash, keyFingerprint: golden.KeyFingerprint,
                                  code: golden.Code, redirectURI: golden.RedirectURI, codeChallenge: golden.CodeChallenge,
                                  state: golden.State, expiresAt: golden.ExpiresAt)
    let message = try! proof.message(publicKey: publicKey, now: now)
    check(message.base64EncodedString() == golden.MessageBase64, "independent monolith vector has identical signed bytes")
    check(scope.storageKey == "08eba32e93e89b49cbed7c040f8de7dba724391c4015de3baaa78d9dad03957f", "stable environment key domain")
    let fixturePrivateKey = try! P256.Signing.PrivateKey(rawRepresentation: Data(repeating: 0, count: 31) + Data([1]))
    check(fixturePrivateKey.publicKey.x963Representation == publicKey, "synthetic fixture public key")
    let fixtureSignature = try! fixturePrivateKey.signature(for: message)
    check(try! installationPublicKey(publicKey).isValidSignature(fixtureSignature, for: message), "typed SHA256 DER verifies")
    var changedMessage = message; changedMessage[changedMessage.count - 1] ^= 1
    check(!(try! installationPublicKey(publicKey).isValidSignature(fixtureSignature, for: changedMessage)), "proof binds expiry bytes")

    final class Keys: InstallationKeyStore {
        let key: P256.Signing.PrivateKey
        let now: Date
        var calls = [String]()
        init(key: P256.Signing.PrivateKey, now: Date) { self.key = key; self.now = now }
        func ensureKey(scope: InstallationScope) throws -> Data {
            calls.append("key:" + scope.storageKey); return key.publicKey.x963Representation
        }
        func signProof(_ proof: InstallationProof) throws -> Data {
            calls.append("proof")
            return try key.signature(for: proof.message(publicKey: key.publicKey.x963Representation, now: now)).derRepresentation
        }
    }
    let keys = Keys(key: fixturePrivateKey, now: now)
    let credentials = FakeStore()
    var confirmations = 0
    let handler = RequestHandler(preflight: {}, store: credentials, confirm: { confirmations += 1 }, installationKeys: keys, now: { now })
    let scopeObject: [String: Any] = ["resource": golden.Resource, "client_id": golden.ClientID]
    let proofObject: [String: Any] = ["resource": golden.Resource, "client_id": golden.ClientID,
        "installation_intent": golden.IntentID, "challenge_hash": golden.ChallengeHash,
        "key_fingerprint": golden.KeyFingerprint, "code": golden.Code, "redirect_uri": golden.RedirectURI,
        "code_challenge": golden.CodeChallenge, "state": golden.State, "expires_at": golden.ExpiresAt]
    func request(_ operation: String, _ payload: [String: Any], interactive: Bool = true, key: String? = nil) -> Data {
        let value = try! JSONSerialization.data(withJSONObject: payload)
        return try! JSONSerialization.data(withJSONObject: ["version": 1, "operation": operation,
            "key": key ?? scope.storageKey, "interactive": interactive, "value": value.base64EncodedString()])
    }
    let first = handler.handle(request("installation-key", scopeObject))
    let second = handler.handle(request("installation-key", scopeObject))
    check(first.ok && first.data?.publicKey == golden.PublicKey && first.data?.keyFingerprint == golden.KeyFingerprint, "public key only")
    check(second.ok && second.data?.publicKey == first.data?.publicKey && confirmations == 2, "fresh confirmation and stable key lookup")
    check(first.data?.value == nil && first.data?.present == nil && first.data?.signature == nil, "key response has no credential")
    let signed = handler.handle(request("installation-proof", proofObject))
    check(signed.ok && signed.data?.algorithm == installationAlgorithm, "typed proof supported")
    let signature = (70...72).compactMap { try? installationDecodeURL(signed.data?.signature ?? "", count: $0) }.first ?? Data()
    // DER signatures can be 70, 71 or 72 bytes; parse and verify the actual result.
    let actualSignature = try? P256.Signing.ECDSASignature(derRepresentation: signature)
    check(actualSignature != nil && fixturePrivateKey.publicKey.isValidSignature(actualSignature!, for: message), "proof uses valid ASN1 DER")
    check(credentials.calls.isEmpty, "installation operations never load/save/delete OAuth credentials")
    let before = keys.calls.count
    for operation in ["installation-key", "installation-proof"] {
        let payload = operation == "installation-key" ? scopeObject : proofObject
        check(handler.handle(request(operation, payload, interactive: false)).error?.code == "INTERACTION_REQUIRED", "no-input key/proof denied")
        let refused = RequestHandler(preflight: { throw HelperFailure.unavailable }, store: credentials,
                                     confirm: { confirmations += 1 }, installationKeys: keys, now: { now })
        check(refused.handle(request(operation, payload)).error?.code == "UNAVAILABLE", "signature before key access")
    }
    check(keys.calls.count == before && confirmations == 2, "refusal has no key or confirmation side effect")
    let cancelling = RequestHandler(preflight: {}, store: credentials, confirm: { throw HelperFailure.cancelled }, installationKeys: keys, now: { now })
    check(cancelling.handle(request("installation-key", scopeObject)).error?.code == "CANCELLED", "confirmation cancellation before key")
    check(keys.calls.count == before, "cancel cannot create key")
    for (field, value): (String, Any) in [("client_id", "other"), ("resource", "http://evil.example"),
        ("resource", golden.Resource + "/"), ("state", "short"), ("installation_intent", "bad"),
        ("code_challenge", golden.CodeChallenge + "="), ("challenge_hash", golden.ChallengeHash.uppercased()),
        ("redirect_uri", "http://localhost:49152/callback"), ("redirect_uri", golden.RedirectURI + "?x=1"),
        ("redirect_uri", "http://127.0.0.1:0/callback"), ("expires_at", golden.ExpiresAt - 300),
        ("expires_at", NSNull()), ("code", String(repeating: "c", count: 257)), ("scope", "write")] {
        var payload = proofObject; payload[field] = value
        check(handler.handle(request("installation-proof", payload)).error?.code == "INVALID_REQUEST", "invalid proof refused: " + field)
    }
    check(keys.calls.count == before, "invalid proof never reaches key store")
    var otherKey = proofObject; otherKey["key_fingerprint"] = String(repeating: "0", count: 64)
    check(handler.handle(request("installation-proof", otherKey)).error?.code == "INVALID_REQUEST", "key fingerprint mismatch cannot sign")
    check(handler.handle(request("installation-key", scopeObject, key: String(repeating: "0", count: 64))).error?.code == "INVALID_REQUEST", "environment scope cannot use another key")
    for raw in ["{\"a\":1,\"a\":2}", "{\"a\":1,\"\\u0061\":2}", "{\"a\":{}}", "{\"a\":[]}", "{\"a\":1,}", "{} {}"] {
        do { _ = try flatJSONObject(Data(raw.utf8)); check(false, "ambiguous flat JSON rejected") } catch {}
    }
    check((try? flatJSONObject(Data("{\"escaped\\\"key\":\"brace{comma,colon:quote\\\"\\\\\"}".utf8))) != nil, "string punctuation remains valid")
    for origin in ["https://api.example.test:443", "http://127.0.0.1:49152", "http://[::1]:49152"] {
        check((try? InstallationScope(resource: origin, clientID: "riido-cli").validate()) != nil, "allowed exact origin")
    }
    // Export only synthetic fixture signature to let Go independently verify it.
    if CommandLine.arguments.count > 2 {
        let evidence = ["message": message.base64EncodedString(), "signature": installationBase64URL(signature), "public_key": golden.PublicKey]
        try! JSONSerialization.data(withJSONObject: evidence, options: [.sortedKeys]).write(to: URL(fileURLWithPath: CommandLine.arguments[2]))
    }
}
