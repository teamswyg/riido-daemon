"""Private, same-user Unix IPC for the manual ended-run QA probe."""

import ctypes
import json
import os
from pathlib import Path
import resource
import socket
import stat
import struct
import sys
import time


class Failure(Exception):
    """Only fixed local reason codes belong in a report."""


def require(value, reason):
    if not value:
        raise Failure(reason)


def disable_core_dumps():
    resource.setrlimit(resource.RLIMIT_CORE, (0, 0))


def private_path(path, existing=False):
    path = Path(path)
    require(path.is_absolute() and len(os.fsencode(path)) < 100, "INVALID_HANDOFF_PATH")
    parent = path.parent
    info = parent.lstat()
    require(not parent.is_symlink() and parent.resolve() == parent,
            "HANDOFF_PARENT_SYMLINK")
    require(stat.S_ISDIR(info.st_mode) and info.st_uid == os.getuid()
            and stat.S_IMODE(info.st_mode) == 0o700, "HANDOFF_PARENT_NOT_PRIVATE")
    if existing:
        info = path.lstat()
        require(stat.S_ISSOCK(info.st_mode) and info.st_uid == os.getuid()
                and stat.S_IMODE(info.st_mode) == 0o600, "HANDOFF_SOCKET_NOT_PRIVATE")
    else:
        require(not path.exists() and not path.is_symlink(), "HANDOFF_ALREADY_EXISTS")
    return str(path)


def same_user_peer(conn):
    if sys.platform == "darwin":
        uid, gid = ctypes.c_uint(), ctypes.c_uint()
        result = ctypes.CDLL(None, use_errno=True).getpeereid(conn.fileno(), ctypes.byref(uid), ctypes.byref(gid))
        require(result == 0, "PEER_ID_UNAVAILABLE")
        peer_uid = uid.value
    elif sys.platform.startswith("linux"):
        _, peer_uid, _ = struct.unpack("3i", conn.getsockopt(socket.SOL_SOCKET, socket.SO_PEERCRED, 12))
    else:
        raise Failure("UNSUPPORTED_PLATFORM")
    require(peer_uid == os.getuid(), "PEER_USER_MISMATCH")


def send_json(conn, value):
    conn.sendall(json.dumps(value, separators=(",", ":")).encode() + b"\n")


def remaining(deadline, limit):
    left = min(limit, deadline - time.monotonic())
    require(left > 0, "PROBE_TIMEOUT")
    return left


def read_json(conn, max_bytes=65536, deadline=None):
    deadline = min(deadline or float("inf"), time.monotonic() + (conn.gettimeout() or 5))
    raw = bytearray()
    while b"\n" not in raw:
        conn.settimeout(remaining(deadline, 40))
        block = conn.recv(min(8192, max_bytes + 1 - len(raw)))
        require(time.monotonic() <= deadline, "PROBE_TIMEOUT")
        require(bool(block), "IPC_EARLY_CLOSE")
        raw.extend(block)
        require(len(raw) <= max_bytes, "IPC_RESPONSE_TOO_LARGE")
    try:
        value = json.loads(raw)
    except (ValueError, UnicodeError):
        raise Failure("IPC_INVALID_JSON") from None
    require(isinstance(value, dict), "IPC_INVALID_OBJECT")
    return value


def daemon_status(path, deadline=None):
    deadline = min(deadline or float("inf"), time.monotonic() + 5)
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
        conn.settimeout(remaining(deadline, 5))
        conn.connect(path)
        same_user_peer(conn)
        send_json(conn, {"method": "status"})
        return read_json(conn, 2 * 1024 * 1024, deadline)
