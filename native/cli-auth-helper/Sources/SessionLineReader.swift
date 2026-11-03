import Foundation
import Darwin

// Reads newline-delimited requests from stdin, each at most 64 KiB. It uses
// read(2), which returns as soon as a request line is available; FileHandle's
// read(upToCount:) waits for the full count or EOF and would stall a session.
struct SessionLineReader {
    var read: (Int) throws -> Data? = SessionLineReader.readStandardInput

    static func readStandardInput(_ count: Int) throws -> Data? {
        var buffer = [UInt8](repeating: 0, count: count)
        while true {
            let n = Darwin.read(STDIN_FILENO, &buffer, count)
            if n > 0 { return Data(buffer[0..<n]) }
            if n == 0 { return nil }
            if errno != EINTR { throw HelperFailure.invalid }
        }
    }
    private var buffer = Data()
    private var finished = false

    init(read: ((Int) throws -> Data?)? = nil) {
        if let read { self.read = read }
    }

    mutating func next() throws -> Data? {
        while true {
            if let newline = buffer.firstIndex(of: 10) {
                let line = buffer[buffer.startIndex..<newline]
                buffer.removeSubrange(buffer.startIndex...newline)
                guard line.count <= 65536 else { throw HelperFailure.invalid }
                return Data(line)
            }
            guard buffer.count <= 65536 else { throw HelperFailure.invalid }
            if finished {
                if buffer.isEmpty { return nil }
                throw HelperFailure.invalid
            }
            guard let part = try read(4096), !part.isEmpty else { finished = true; continue }
            buffer.append(part)
        }
    }
}
