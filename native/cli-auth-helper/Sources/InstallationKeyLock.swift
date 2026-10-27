import Foundation
import Darwin

// This empty-file advisory lock coordinates trusted helper processes, including
// direct private-protocol callers. OS presence/keychain policy remain the
// credential boundary; a malicious peer with this UID can still deny service.
struct InstallationKeyLock {
    var directory: () throws -> Int32 = installationLockDirectory
    var timeout: TimeInterval = 5

    func acquire(key: String) throws -> InstallationKeyReservation {
        guard key.utf8.count == 64, key.utf8.allSatisfy({ (48...57).contains($0) || (97...102).contains($0) }) else {
            throw HelperFailure.invalid
        }
        let directoryFD = try directory()
        let name = key + ".lock"
        let fd = openat(directoryFD, name, O_CREAT | O_RDWR | O_CLOEXEC | O_NOFOLLOW, 0o600)
        guard fd >= 0 else { close(directoryFD); throw HelperFailure.unavailable }
        let release = { _ = flock(fd, LOCK_UN); close(fd); close(directoryFD) }
        func checkIdentity() throws {
            var descriptor = stat(), entry = stat()
            guard fstat(fd, &descriptor) == 0,
                  fstatat(directoryFD, name, &entry, AT_SYMLINK_NOFOLLOW) == 0,
                  descriptor.st_uid == getuid(), descriptor.st_mode & S_IFMT == S_IFREG,
                  descriptor.st_mode & 0o077 == 0, descriptor.st_nlink == 1,
                  descriptor.st_dev == entry.st_dev, descriptor.st_ino == entry.st_ino,
                  entry.st_mode & S_IFMT == S_IFREG else { throw HelperFailure.unavailable }
        }
        do {
            try checkIdentity()
            let deadline = ProcessInfo.processInfo.systemUptime + timeout
            while flock(fd, LOCK_EX | LOCK_NB) != 0 {
                guard errno == EWOULDBLOCK || errno == EAGAIN || errno == EINTR else { throw HelperFailure.unavailable }
                guard ProcessInfo.processInfo.systemUptime < deadline else { throw HelperFailure.timedOut }
                Thread.sleep(forTimeInterval: 0.02)
            }
            try checkIdentity()
            return InstallationKeyReservation(check: checkIdentity, release: release)
        } catch { release(); throw error }
    }
}

private func installationLockDirectory() throws -> Int32 {
    // Use the operating-system account's home, never a caller-controlled HOME
    // or TMPDIR override. Open each owned directory without following symlinks.
    guard let account = getpwuid(getuid()), let home = account.pointee.pw_dir else { throw HelperFailure.unavailable }
    let homeFD = open(home, O_RDONLY | O_DIRECTORY | O_CLOEXEC | O_NOFOLLOW)
    guard homeFD >= 0 else { throw HelperFailure.unavailable }
    defer { close(homeFD) }
    var attributes = stat()
    guard fstat(homeFD, &attributes) == 0, attributes.st_uid == getuid(),
          attributes.st_mode & 0o022 == 0 else { throw HelperFailure.unavailable }
    return try installationLockDirectory(homeFD: homeFD)
}

func installationLockDirectory(homeFD: Int32) throws -> Int32 {
    let name = ".riido-cli-installation-locks"
    if mkdirat(homeFD, name, 0o700) != 0 && errno != EEXIST { throw HelperFailure.unavailable }
    let fd = openat(homeFD, name, O_RDONLY | O_DIRECTORY | O_CLOEXEC | O_NOFOLLOW)
    guard fd >= 0 else { throw HelperFailure.unavailable }
    var attributes = stat()
    guard fstat(fd, &attributes) == 0, attributes.st_uid == getuid(),
          attributes.st_mode & S_IFMT == S_IFDIR, attributes.st_mode & 0o077 == 0 else {
        close(fd); throw HelperFailure.unavailable
    }
    return fd
}
