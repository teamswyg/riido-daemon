#if !CLI_AUTH_HELPER_TESTING
import Foundation

@main
struct HelperMain {
    static func main() {
        let handler = RequestHandler(preflight: verifySigningPreflight, store: ProtectedKeychainStore(),
                                     confirm: { try UserPresenceVerifier().confirm() },
                                     installationKeys: ProtectedInstallationKeyStore())
        let input: Data
        do { input = try readBoundedInput() }
        catch { write(HelperResponse(ok: false, error: .invalid)); return }
        write(handler.handle(input))
    }
    private static func readBoundedInput() throws -> Data {
        var input = Data()
        while input.count <= 65536 {
            guard let part = try FileHandle.standardInput.read(upToCount: min(4096, 65537 - input.count)), !part.isEmpty else { return input }
            input.append(part)
        }
        throw HelperFailure.invalid
    }
    private static func write(_ response: HelperResponse) {
        guard var data = try? JSONEncoder().encode(response) else { return }
        data.append(10)
        FileHandle.standardOutput.write(data)
    }
}
#endif
