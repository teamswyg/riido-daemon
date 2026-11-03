import Foundation

// Reads newline-delimited requests from stdin, each at most 64 KiB.
struct SessionLineReader {
    var read: (Int) throws -> Data? = { try FileHandle.standardInput.read(upToCount: $0) }
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
