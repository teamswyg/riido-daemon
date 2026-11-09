import Foundation
import Darwin

// Read inside the signed helper. No environment variable, file, subprocess or
// caller-provided boot identifier participates in the witness.
func currentBootSessionID() throws -> String {
    var count = 0
    guard sysctlbyname("kern.bootsessionuuid", nil, &count, nil, 0) == 0,
          count == 37 else { throw HelperFailure.unavailable }
    var bytes = [UInt8](repeating: 0, count: count)
    let status = bytes.withUnsafeMutableBytes { buffer in
        sysctlbyname("kern.bootsessionuuid", buffer.baseAddress, &count, nil, 0)
    }
    guard status == 0, count == 37, bytes.last == 0,
          let value = String(bytes: bytes.dropLast(), encoding: .utf8) else { throw HelperFailure.unavailable }
    return try canonicalBootSessionID(value)
}

func canonicalBootSessionID(_ value: String) throws -> String {
    guard value.utf8.count == 36, let uuid = UUID(uuidString: value),
          uuid.uuidString.lowercased() == value.lowercased(),
          uuid.uuidString != "00000000-0000-0000-0000-000000000000" else { throw HelperFailure.unavailable }
    return uuid.uuidString.lowercased()
}

func scopedBootSessionDigest(_ proof: DeviceBootstrapProof, bootID: String) throws -> String {
    let canonical = try canonicalBootSessionID(bootID)
    return installationHex(installationDigest(installationFrame([
        "riido-cli-boot-session.v1", proof.link.resource, proof.link.clientID,
        proof.link.keyFingerprint, canonical].map { Data($0.utf8) })))
}
