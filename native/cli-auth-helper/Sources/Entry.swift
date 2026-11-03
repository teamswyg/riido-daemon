#if !CLI_AUTH_HELPER_TESTING
import Foundation

@main
struct HelperMain {
    static func main() {
        if CommandLine.arguments.dropFirst() == ["--session"] { return serveSession() }
        let handler = RequestHandler(preflight: verifySigningPreflight, store: ProtectedKeychainStore(),
                                     confirm: { try UserPresenceVerifier().confirm() },
                                     installationKeys: ProtectedInstallationKeyStore())
        let input: Data
        do { input = try readBoundedInput() }
        catch { write(HelperResponse(ok: false, error: .invalid)); return }
        write(handler.handle(input))
    }
    // One CLI command, one helper process, one OS confirmation. Requests are
    // newline-delimited flat JSON; the session ends on stdin EOF, after
    // sessionRequestLimit requests or after sessionLifetime, whichever first.
    static let sessionRequestLimit = 64
    static let sessionLifetime: TimeInterval = 600

    private static func serveSession() {
        let session = AuthenticationSession()
        currentAuthenticationSession = session
        defer { session.invalidate(); currentAuthenticationSession = nil }
        var handler = RequestHandler(preflight: verifySigningPreflight, store: ProtectedKeychainStore(),
                                     confirm: { try session.ensureVerified() },
                                     installationKeys: ProtectedInstallationKeyStore())
        handler.beforeProtected = { request in
            guard request.interactive else { throw HelperFailure.interactionRequired }
            try session.ensureVerified()
        }
        let deadline = Date().addingTimeInterval(sessionLifetime)
        var reader = SessionLineReader()
        for _ in 0..<sessionRequestLimit {
            let line: Data
            do {
                guard let next = try reader.next() else { return }
                line = next
            } catch { write(HelperResponse(ok: false, error: .invalid)); return }
            guard Date() < deadline else { write(HelperResponse(ok: false, error: .unavailable)); return }
            write(handler.handle(line))
        }
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
