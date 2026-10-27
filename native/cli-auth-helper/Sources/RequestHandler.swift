import Foundation

struct RequestHandler {
    let preflight: () throws -> Void
    let store: CredentialStore
    let confirm: () throws -> Void
    var installationKeys: InstallationKeyStore = UnavailableInstallationKeyStore()
    var now: () -> Date = Date.init
    func handle(_ input: Data) -> HelperResponse {
        do {
            let request = try decode(input)
            try preflight()
            let data: HelperData
            switch request.operation {
            case "status":
                data = HelperData(present: try store.status(key: request.key, interactive: request.interactive))
            case "load":
                let value = try store.load(key: request.key, interactive: request.interactive)
                guard !value.isEmpty, value.count <= 32768 else { throw HelperFailure.unavailable }
                data = HelperData(present: true, value: value.base64EncodedString())
            case "save":
                guard let text = request.value, let value = Data(base64Encoded: text) else { throw HelperFailure.invalid }
                try store.save(key: request.key, value: value, interactive: request.interactive)
                data = HelperData(present: true)
            case "delete":
                try store.delete(key: request.key, interactive: request.interactive)
                data = HelperData(present: false)
            case "confirm":
                guard request.interactive else { throw HelperFailure.interactionRequired }
                try confirm()
                data = HelperData(confirmed: true)
            case "installation-key":
                guard request.interactive else { throw HelperFailure.interactionRequired }
                let scope = try InstallationScope.decode(installationValue(request))
                guard scope.storageKey == request.key else { throw HelperFailure.invalid }
                try confirm()
                let publicKey = try installationKeys.ensureKey(scope: scope)
                _ = try installationPublicKey(publicKey)
                data = HelperData(algorithm: installationAlgorithm, publicKey: installationBase64URL(publicKey),
                                  keyFingerprint: installationHex(installationDigest(publicKey)))
            case "installation-proof":
                guard request.interactive else { throw HelperFailure.interactionRequired }
                let proof = try InstallationProof.decode(installationValue(request), now: now())
                guard proof.scope.storageKey == request.key else { throw HelperFailure.invalid }
                let signature = try installationKeys.signProof(proof)
                guard !signature.isEmpty, signature.count <= 72 else { throw HelperFailure.unavailable }
                data = HelperData(algorithm: installationAlgorithm, signature: installationBase64URL(signature))
            default: throw HelperFailure.invalid
            }
            return HelperResponse(ok: true, data: data)
        } catch let failure as HelperFailure {
            return HelperResponse(ok: false, error: failure)
        } catch {
            return HelperResponse(ok: false, error: .unavailable)
        }
    }
    private func decode(_ input: Data) throws -> HelperRequest {
        let object = try flatJSONObject(input)
        guard
              Set(object.keys).isSubset(of: ["version", "operation", "key", "interactive", "value"]),
              let request = try? JSONDecoder().decode(HelperRequest.self, from: input),
              request.version == 1,
              ["status", "load", "save", "delete", "confirm", "installation-key", "installation-proof"].contains(request.operation),
              (16...128).contains(request.key.utf8.count),
              request.key.utf8.allSatisfy({ byte in
                  (65...90).contains(byte) || (97...122).contains(byte) || (48...57).contains(byte) || byte == 45 || byte == 95
              }) else { throw HelperFailure.invalid }
        if ["installation-key", "installation-proof"].contains(request.operation) {
            let bytes = try installationValue(request)
            if request.operation == "installation-key" {
                guard try InstallationScope.decode(bytes).storageKey == request.key else { throw HelperFailure.invalid }
            } else {
                guard try InstallationProof.decode(bytes, now: now()).scope.storageKey == request.key else { throw HelperFailure.invalid }
            }
        } else if request.operation == "save" {
            guard let text = request.value, let value = Data(base64Encoded: text),
                  !value.isEmpty, value.count <= 32768 else { throw HelperFailure.invalid }
        } else if object.keys.contains("value") { throw HelperFailure.invalid }
        return request
    }
    private func installationValue(_ request: HelperRequest) throws -> Data {
        guard let value = request.value, let bytes = Data(base64Encoded: value),
              !bytes.isEmpty, bytes.count <= 8192, bytes.base64EncodedString() == value else { throw HelperFailure.invalid }
        return bytes
    }
}
