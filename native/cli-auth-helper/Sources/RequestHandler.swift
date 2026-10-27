import Foundation

struct RequestHandler {
    let preflight: () throws -> Void
    let store: CredentialStore
    let confirm: () throws -> Void
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
        guard !input.isEmpty, input.count <= 65536,
              let object = try? JSONSerialization.jsonObject(with: input) as? [String: Any],
              Set(object.keys).isSubset(of: ["version", "operation", "key", "interactive", "value"]),
              let request = try? JSONDecoder().decode(HelperRequest.self, from: input),
              request.version == 1,
              ["status", "load", "save", "delete", "confirm"].contains(request.operation),
              (16...128).contains(request.key.utf8.count),
              request.key.utf8.allSatisfy({ byte in
                  (65...90).contains(byte) || (97...122).contains(byte) || (48...57).contains(byte) || byte == 45 || byte == 95
              }) else { throw HelperFailure.invalid }
        if request.operation == "save" {
            guard let text = request.value, let value = Data(base64Encoded: text),
                  !value.isEmpty, value.count <= 32768 else { throw HelperFailure.invalid }
        } else if object.keys.contains("value") { throw HelperFailure.invalid }
        return request
    }
}
