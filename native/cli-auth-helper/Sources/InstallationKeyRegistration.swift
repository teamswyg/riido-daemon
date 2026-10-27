import Foundation

struct InstallationKeyReservation {
    let check: () throws -> Void
    let release: () -> Void
}

// Keychain application tags are not unique keys: application labels also
// participate in the primary key. Registration must serialize find/create and
// reject ambiguous matches instead of choosing an arbitrary permanent key.
struct InstallationKeyRegistration<Key> {
    let reserve: () throws -> InstallationKeyReservation
    let load: () throws -> [Key]
    let create: () throws -> Key
    let publicBytes: (Key) throws -> Data

    func ensure() throws -> Data {
        let reservation = try reserve()
        defer { reservation.release() }
        try reservation.check()
        let existing = try load()
        guard existing.count <= 1 else { throw HelperFailure.unavailable }
        let result: Data
        if let key = existing.first { result = try publicBytes(key) }
        else {
            result = try publicBytes(create())
            let persisted = try load()
            guard persisted.count == 1, try publicBytes(persisted[0]) == result else { throw HelperFailure.unavailable }
        }
        try reservation.check()
        return result
    }
}
