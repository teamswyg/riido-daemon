import Foundation

extension InstallationKeyStore {
    // Optional implementations cannot substitute key creation for this lookup.
    func existingKey(scope: InstallationScope) throws -> Data { throw HelperFailure.unavailable }
}
