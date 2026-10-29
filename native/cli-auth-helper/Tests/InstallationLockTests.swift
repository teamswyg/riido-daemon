import Foundation
import Darwin

private func fixtureLockDirectory(_ home: String) throws -> Int32 {
    let fd = open(home, O_RDONLY | O_DIRECTORY | O_CLOEXEC | O_NOFOLLOW)
    guard fd >= 0 else { throw HelperFailure.unavailable }
    defer { close(fd) }
    return try installationLockDirectory(homeFD: fd)
}

func installationLockProbe() -> Int32 {
    guard CommandLine.arguments.count == 4 else { return 2 }
    let home = CommandLine.arguments[2], key = CommandLine.arguments[3]
    do {
        let held = try InstallationKeyLock(directory: { try fixtureLockDirectory(home) }, timeout: 0.15).acquire(key: key)
        defer { held.release() }
        try held.check()
        return 0
    } catch { return 10 }
}

func runInstallationLockTests() {
    let root = FileManager.default.temporaryDirectory.appendingPathComponent("riido-install-lock-" + UUID().uuidString)
    try! FileManager.default.createDirectory(at: root, withIntermediateDirectories: false, attributes: [.posixPermissions: 0o700])
    defer { try? FileManager.default.removeItem(at: root) }
    let scope = String(repeating: "1", count: 64)
    let lock = InstallationKeyLock(directory: { try fixtureLockDirectory(root.path) }, timeout: 0.15)
    func child(_ key: String) -> Int32 {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: CommandLine.arguments[0])
        process.arguments = ["--installation-lock-probe", root.path, key]
        process.standardOutput = Pipe(); process.standardError = Pipe()
        do { try process.run() } catch { return 2 }
        process.waitUntilExit()
        return process.terminationStatus
    }
    do {
        let held = try lock.acquire(key: scope)
        check(child(scope) == 10, "separate helper cannot acquire live scope lock")
        check(child(String(repeating: "2", count: 64)) == 0, "different scope can acquire lock")
        held.release()
        check(child(scope) == 0, "scope lock releases across processes")
        let path = root.appendingPathComponent(".riido-cli-installation-locks/" + scope + ".lock")
        let lockBytes = try Data(contentsOf: path)
        check(lockBytes.isEmpty, "lock stores no credential or key material")
        let changed = try lock.acquire(key: scope)
        try FileManager.default.removeItem(at: path)
        try Data().write(to: path)
        try FileManager.default.setAttributes([.posixPermissions: 0o600], ofItemAtPath: path.path)
        do { try changed.check(); check(false, "replaced lock inode must refuse") } catch {}
        changed.release()
        check(FileManager.default.fileExists(atPath: path.path), "release never deletes another lock inode")
        try FileManager.default.setAttributes([.posixPermissions: 0o666], ofItemAtPath: path.path)
        do { let invalid = try lock.acquire(key: scope); invalid.release(); check(false, "writable lock refused") } catch {}
        try FileManager.default.removeItem(at: path)
        let target = root.appendingPathComponent("unrelated")
        try Data().write(to: target)
        try FileManager.default.createSymbolicLink(at: path, withDestinationURL: target)
        do { let invalid = try lock.acquire(key: scope); invalid.release(); check(false, "symlink lock refused") } catch {}
        try FileManager.default.removeItem(at: path)
        try FileManager.default.setAttributes([.posixPermissions: 0o600], ofItemAtPath: target.path)
        try FileManager.default.linkItem(at: target, to: path)
        do { let invalid = try lock.acquire(key: scope); invalid.release(); check(false, "hardlinked lock refused") } catch {}
        do { let invalid = try lock.acquire(key: "../" + scope); invalid.release(); check(false, "path scope refused") } catch {}
        let directory = root.appendingPathComponent(".riido-cli-installation-locks")
        try FileManager.default.removeItem(at: directory)
        try FileManager.default.createSymbolicLink(at: directory, withDestinationURL: root)
        do { let invalid = try lock.acquire(key: scope); invalid.release(); check(false, "symlink directory refused") } catch {}
    } catch { check(false, "native lock fixture unexpected error") }
}
