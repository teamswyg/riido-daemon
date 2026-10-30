import Foundation
import CryptoKit

func runInstallationRepairTests() {
    let privateKey = P256.Signing.PrivateKey()
    let publicKey = privateKey.publicKey.x963Representation
    let fingerprint = installationHex(installationDigest(publicKey))
    let scope = InstallationScope(resource: "https://cli.fixture.invalid", clientID: "riido-cli")
    let payload: [String:Any] = ["resource":scope.resource,"client_id":scope.clientID,
        "request_id":String(repeating:"A",count:43),"key_fingerprint":fingerprint,"recovery_hash":String(repeating:"a",count:64)]
    final class Keys: InstallationKeyStore {
        let key: P256.Signing.PrivateKey
        var calls = 0
        init(_ key:P256.Signing.PrivateKey) { self.key=key }
        func ensureKey(scope:InstallationScope) throws -> Data { check(false,"repair never creates a key");throw HelperFailure.unavailable }
        func signProof(_ proof:InstallationProof) throws -> Data { check(false,"repair never uses OAuth purpose");throw HelperFailure.unavailable }
        func signRepairProof(_ proof:InstallationRepairProof) throws -> Data {
            let message = try proof.message(publicKey:key.publicKey.x963Representation)
            calls += 1
            return try key.signature(for:message).derRepresentation
        }
    }
    func request(_ value:[String:Any],interactive:Bool=true,key:String?=nil) -> Data {
        let body = try! JSONSerialization.data(withJSONObject:value)
        return try! JSONSerialization.data(withJSONObject:["version":1,"operation":"installation-repair-proof","key":key ?? scope.storageKey,"interactive":interactive,"value":body.base64EncodedString()])
    }
    let keys = Keys(privateKey)
    let credentials = FakeStore()
    let handler = RequestHandler(preflight:{},store:credentials,confirm:{},installationKeys:keys)
    let response = handler.handle(request(payload))
    check(response.ok && keys.calls==1,"repair uses typed existing-key signer")
    let proof = try! InstallationRepairProof.decode(JSONSerialization.data(withJSONObject:payload))
    let signatureText = response.data?.signature ?? ""
    let signatureData = (8...72).compactMap{try? installationDecodeURL(signatureText,count:$0)}.first ?? Data()
    let signature = try? P256.Signing.ECDSASignature(derRepresentation:signatureData)
    check(signature != nil && privateKey.publicKey.isValidSignature(signature!,for:try! proof.message(publicKey:publicKey)),"repair DER verifies for exact request")
    check(handler.handle(request(payload,interactive:false)).error?.code=="INTERACTION_REQUIRED","repair requires OS interaction")
    check(!handler.handle(request(payload,key:String(repeating:"0",count:64))).ok,"repair rejects wrong native scope")
    for (field,value) in [("resource","http://localhost"),("client_id","other"),("request_id","invalid"),("recovery_hash",String(repeating:"A",count:64)),("key_fingerprint",String(repeating:"b",count:64)),("extra","arbitrary-sign")] {
        var changed=payload;changed[field]=value
        check(!handler.handle(request(changed)).ok,"repair rejects "+field)
    }
    let raw=try! JSONSerialization.data(withJSONObject:payload)
    let duplicate=String(data:raw,encoding:.utf8)!.dropLast()+",\"client_id\":\"riido-cli\"}"
    check((try? InstallationRepairProof.decode(Data(duplicate.utf8)))==nil,"repair rejects duplicate fields")
    check(keys.calls==1 && credentials.calls.isEmpty,"invalid repair never signs or reads credentials")
    let blocked=RequestHandler(preflight:{throw HelperFailure.unavailable},store:credentials,confirm:{},installationKeys:keys)
    check(blocked.handle(request(payload)).error?.code=="UNAVAILABLE" && keys.calls==1,"repair signature preflight precedes key access")
    let old=RequestHandler(preflight:{},store:credentials,confirm:{})
    check(old.handle(request(payload)).error?.code=="UNAVAILABLE","older signer cannot fall back")
    let fixture=try! JSONSerialization.jsonObject(with:Data(contentsOf:URL(fileURLWithPath:CommandLine.arguments[1]))) as! [String:Any]
    let fixtureKey=try! installationDecodeURL(fixture["PublicKey"] as! String,count:65)
    let canonical=InstallationRepairProof(resource:scope.resource,clientID:scope.clientID,requestID:String(repeating:"A",count:43),keyFingerprint:fixture["KeyFingerprint"] as! String,recoveryHash:String(repeating:"a",count:64))
    check(installationHex(installationDigest(try! canonical.message(publicKey:fixtureKey)))=="0c57a0f38699744f88d32a8aaff7b224426374a3a9b3c8399731b023a6d038de","repair bytes match Go/server golden")
}
