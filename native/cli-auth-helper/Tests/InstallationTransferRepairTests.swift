import Foundation
import CryptoKit

func runInstallationTransferRepairTests() {
    let privateKey = P256.Signing.PrivateKey()
    let publicKey = privateKey.publicKey.x963Representation
    let fingerprint = installationHex(installationDigest(publicKey))
    let scope = InstallationScope(resource: "https://cli.fixture.invalid", clientID: "riido-cli")
    let payload: [String:Any] = ["resource":scope.resource,"client_id":scope.clientID,
        "request_id":String(repeating:"A",count:43),"original_request_id":String(repeating:"B",count:42)+"A","key_fingerprint":fingerprint,"recovery_hash":String(repeating:"a",count:64)]
    final class Keys: InstallationKeyStore {
        let key: P256.Signing.PrivateKey
        var calls = 0
        init(_ key:P256.Signing.PrivateKey) { self.key=key }
        func ensureKey(scope:InstallationScope) throws -> Data { check(false,"repair never creates a key");throw HelperFailure.unavailable }
        func signProof(_ proof:InstallationProof) throws -> Data { check(false,"repair never uses OAuth purpose");throw HelperFailure.unavailable }
        func signTransferRepairProof(_ proof:InstallationTransferRepairProof) throws -> Data {
            let message = try proof.message(publicKey:key.publicKey.x963Representation)
            calls += 1
            return try key.signature(for:message).derRepresentation
        }
    }
    func request(_ value:[String:Any],interactive:Bool=true,key:String?=nil) -> Data {
        let body = try! JSONSerialization.data(withJSONObject:value)
        return try! JSONSerialization.data(withJSONObject:["version":1,"operation":"installation-transfer-repair-proof","key":key ?? scope.storageKey,"interactive":interactive,"value":body.base64EncodedString()])
    }
    let keys = Keys(privateKey)
    let credentials = FakeStore()
    let handler = RequestHandler(preflight:{},store:credentials,confirm:{},installationKeys:keys)
    let response = handler.handle(request(payload))
    check(response.ok && keys.calls==1,"repair uses typed existing-key signer")
    let proof = try! InstallationTransferRepairProof.decode(JSONSerialization.data(withJSONObject:payload))
    let signatureText = response.data?.signature ?? ""
    let signatureData = (8...72).compactMap{try? installationDecodeURL(signatureText,count:$0)}.first ?? Data()
    let signature = try? P256.Signing.ECDSASignature(derRepresentation:signatureData)
    check(signature != nil && privateKey.publicKey.isValidSignature(signature!,for:try! proof.message(publicKey:publicKey)),"repair DER verifies for exact request")
    let initial = InstallationRepairProof(resource:scope.resource,clientID:scope.clientID,requestID:proof.requestID,keyFingerprint:fingerprint,recoveryHash:proof.recoveryHash)
    check(signature != nil && !privateKey.publicKey.isValidSignature(signature!,for:try! initial.message(publicKey:publicKey)),"transfer repair cannot authorize initial repair")
    var swapped=payload;swapped["original_request_id"]=String(repeating:"C",count:42)+"A"
    let changed=try! InstallationTransferRepairProof.decode(JSONSerialization.data(withJSONObject:swapped))
    check(signature != nil && !privateKey.publicKey.isValidSignature(signature!,for:try! changed.message(publicKey:publicKey)),"transfer repair binds original request")
    check(handler.handle(request(payload,interactive:false)).error?.code=="INTERACTION_REQUIRED","repair requires OS interaction")
    check(!handler.handle(request(payload,key:String(repeating:"0",count:64))).ok,"repair rejects wrong native scope")
    for (field,value) in [("resource","http://localhost"),("client_id","other"),("request_id","invalid"),("original_request_id","invalid"),("original_request_id",String(repeating:"A",count:43)),("recovery_hash",String(repeating:"A",count:64)),("key_fingerprint",String(repeating:"b",count:64)),("extra","arbitrary-sign")] {
        var changed=payload;changed[field]=value
        check(!handler.handle(request(changed)).ok,"repair rejects "+field)
    }
    let raw=try! JSONSerialization.data(withJSONObject:payload)
    let duplicate=String(data:raw,encoding:.utf8)!.dropLast()+",\"client_id\":\"riido-cli\"}"
    check((try? InstallationTransferRepairProof.decode(Data(duplicate.utf8)))==nil,"repair rejects duplicate fields")
    check(keys.calls==1 && credentials.calls.isEmpty,"invalid repair never signs or reads credentials")
    let blocked=RequestHandler(preflight:{throw HelperFailure.unavailable},store:credentials,confirm:{},installationKeys:keys)
    check(blocked.handle(request(payload)).error?.code=="UNAVAILABLE" && keys.calls==1,"repair signature preflight precedes key access")
    let old=RequestHandler(preflight:{},store:credentials,confirm:{})
    check(old.handle(request(payload)).error?.code=="UNAVAILABLE","older signer cannot fall back")
    let fixture=try! JSONSerialization.jsonObject(with:Data(contentsOf:URL(fileURLWithPath:CommandLine.arguments[1]))) as! [String:Any]
    let fixtureKey=try! installationDecodeURL(fixture["PublicKey"] as! String,count:65)
    let canonical=InstallationTransferRepairProof(resource:scope.resource,clientID:scope.clientID,requestID:String(repeating:"A",count:43),originalRequestID:String(repeating:"B",count:42)+"A",keyFingerprint:fixture["KeyFingerprint"] as! String,recoveryHash:String(repeating:"a",count:64))
    check(installationHex(installationDigest(try! canonical.message(publicKey:fixtureKey)))=="0a890fbf1ba54cc46de8773baf57950128f39b9652ec99be4d6798860a584552","repair bytes match Go/server golden")
}
