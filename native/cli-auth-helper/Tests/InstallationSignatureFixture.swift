import Foundation
import CryptoKit

func installationFixtureSignature(_ encoded: String) -> Data {
    // Unpadded base64url encodes six bits per byte; DER parsing below decides validity.
    (try? installationDecodeURL(encoded, count: encoded.utf8.count * 3 / 4)) ?? Data()
}

func checkShortInstallationSignature(publicKey: P256.Signing.PublicKey, message: Data) {
    // Synthetic key 1, signed over the independent installation-proof-v1 golden message.
    let bytes = Data(base64Encoded: "MEMCH3b/ijg4jZBZlUB62QMZ6Y6btbK+At90wP8WFwsPz2wCIGqVQ7QlNud7M1K2wLuDL3b9FSz3fDcEC+t8etjQPvXN")!
    let parsed = try! P256.Signing.ECDSASignature(derRepresentation: bytes)
    check(bytes.count == 69 && publicKey.isValidSignature(parsed, for: message), "short DER fixture independently verifies")
    check(installationFixtureSignature(installationBase64URL(bytes)) == bytes, "proof fixture accepts valid 69-byte DER")
    check(installationFixtureSignature(installationBase64URL(bytes) + "=").isEmpty, "proof fixture rejects padded base64url")
    var changedMessage = message; changedMessage[changedMessage.count - 1] ^= 1
    check(!publicKey.isValidSignature(parsed, for: changedMessage), "short DER fixture binds exact message")
}
