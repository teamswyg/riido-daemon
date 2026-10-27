import Foundation

// These private messages contain only primitive values. Foundation's decoder
// otherwise silently accepts duplicate keys, including escaped equivalents.
func flatJSONObject(_ input: Data, maximum: Int = 65536) throws -> [String: Any] {
    guard !input.isEmpty, input.count <= maximum,
          let object = try? JSONSerialization.jsonObject(with: input) as? [String: Any] else {
        throw HelperFailure.invalid
    }
    let bytes = Array(input)
    var index = 0
    func skipSpace() {
        while index < bytes.count && [9, 10, 13, 32].contains(bytes[index]) { index += 1 }
    }
    func stringEnd() throws -> Range<Int> {
        let start = index
        guard index < bytes.count && bytes[index] == 34 else { throw HelperFailure.invalid }
        index += 1
        while index < bytes.count {
            if bytes[index] == 34 { index += 1; return start..<index }
            if bytes[index] == 92 { index += 1 }
            index += 1
        }
        throw HelperFailure.invalid
    }
    skipSpace()
    guard index < bytes.count && bytes[index] == 123 else { throw HelperFailure.invalid }
    index += 1
    var seen = Set<String>()
    skipSpace()
    while index < bytes.count && bytes[index] != 125 {
        let range = try stringEnd()
        guard let key = try? JSONSerialization.jsonObject(with: Data(bytes[range]), options: [.fragmentsAllowed]) as? String,
              seen.insert(key).inserted else { throw HelperFailure.invalid }
        skipSpace()
        guard index < bytes.count && bytes[index] == 58 else { throw HelperFailure.invalid }
        index += 1
        skipSpace()
        guard index < bytes.count else { throw HelperFailure.invalid }
        if bytes[index] == 34 {
            _ = try stringEnd()
        } else {
            while index < bytes.count && bytes[index] != 44 && bytes[index] != 125 {
                guard bytes[index] != 91 && bytes[index] != 123 else { throw HelperFailure.invalid }
                index += 1
            }
        }
        skipSpace()
        if index < bytes.count && bytes[index] == 44 {
            index += 1; skipSpace()
            guard index < bytes.count && bytes[index] != 125 else { throw HelperFailure.invalid }
        }
        else { break }
    }
    guard index < bytes.count && bytes[index] == 125 else { throw HelperFailure.invalid }
    index += 1
    skipSpace()
    guard index == bytes.count, seen.count == object.count else { throw HelperFailure.invalid }
    return object
}
