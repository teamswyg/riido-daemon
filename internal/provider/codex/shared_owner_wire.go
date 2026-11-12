package codex

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1" // RFC 6455 handshake, not a security digest.
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

const sharedWireLimit = 2 * 1024 * 1024

// The official proxy relays raw socket bytes, not JSON lines. This private
// client supports only the uncompressed, bounded WebSocket exchange needed by
// the local Codex control socket. It cannot connect to a network endpoint.
type sharedOwnerWire struct {
	r *bufio.Reader
	w io.Writer
}

func upgradeSharedOwner(stdout io.Reader, stdin io.Writer) (*sharedOwnerWire, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	key := base64.StdEncoding.EncodeToString(nonce[:])
	if _, err := fmt.Fprintf(stdin, "GET / HTTP/1.1\r\nHost: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", key); err != nil {
		return nil, err
	}
	r := bufio.NewReader(io.LimitReader(stdout, sharedWireLimit))
	response, err := http.ReadResponse(r, &http.Request{Method: http.MethodGet})
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // A 101 response has http.NoBody; the upgraded reader remains ours.
	digest := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	if response.StatusCode != http.StatusSwitchingProtocols ||
		!strings.EqualFold(response.Header.Get("Upgrade"), "websocket") ||
		!headerToken(response.Header.Get("Connection"), "upgrade") ||
		response.Header.Get("Sec-WebSocket-Accept") != base64.StdEncoding.EncodeToString(digest[:]) ||
		response.Header.Get("Sec-WebSocket-Extensions") != "" || response.Header.Get("Sec-WebSocket-Protocol") != "" {
		return nil, errors.New("invalid Codex control socket handshake")
	}
	return &sharedOwnerWire{r: r, w: stdin}, nil
}

func headerToken(value, token string) bool {
	for part := range strings.SplitSeq(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func (w *sharedOwnerWire) write(opcode byte, body []byte) error {
	if len(body) > sharedWireLimit || (opcode >= 8 && len(body) > 125) {
		return errors.New("codex control frame too large")
	}
	frame := []byte{0x80 | opcode}
	switch {
	case len(body) < 126:
		frame = append(frame, 0x80|byte(len(body)))
	case len(body) <= 65535:
		frame = append(frame, 0x80|126)
		frame = binary.BigEndian.AppendUint16(frame, uint16(len(body)))
	default:
		frame = append(frame, 0x80|127)
		frame = binary.BigEndian.AppendUint64(frame, uint64(len(body)))
	}
	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return err
	}
	frame = append(frame, mask[:]...)
	for i, b := range body {
		frame = append(frame, b^mask[i%4])
	}
	n, err := w.w.Write(frame)
	if err == nil && n != len(frame) {
		err = io.ErrShortWrite
	}
	return err
}

func (w *sharedOwnerWire) read() ([]byte, error) {
	var message []byte
	fragmented := false
	for range 512 {
		var header [2]byte
		if _, err := io.ReadFull(w.r, header[:]); err != nil {
			return nil, err
		}
		final, opcode := header[0]&0x80 != 0, header[0]&0x0f
		if header[0]&0x70 != 0 || header[1]&0x80 != 0 {
			return nil, errors.New("invalid Codex control frame flags")
		}
		size := uint64(header[1] & 0x7f)
		switch size {
		case 126:
			var length [2]byte
			if _, err := io.ReadFull(w.r, length[:]); err != nil {
				return nil, err
			}
			size = uint64(binary.BigEndian.Uint16(length[:]))
		case 127:
			var length [8]byte
			if _, err := io.ReadFull(w.r, length[:]); err != nil {
				return nil, err
			}
			size = binary.BigEndian.Uint64(length[:])
		}
		if size > sharedWireLimit-uint64(len(message)) || (opcode >= 8 && (!final || size > 125)) {
			return nil, errors.New("codex control frame exceeds bound")
		}
		body := make([]byte, int(size))
		if _, err := io.ReadFull(w.r, body); err != nil {
			return nil, err
		}
		switch opcode {
		case 8:
			return nil, io.EOF
		case 9:
			if err := w.write(10, body); err != nil {
				return nil, err
			}
			continue
		case 10:
			continue
		case 1:
			if fragmented {
				return nil, errors.New("interleaved Codex control messages")
			}
			fragmented = true
		case 0:
			if !fragmented {
				return nil, errors.New("unexpected Codex continuation frame")
			}
		default:
			return nil, errors.New("unsupported Codex control opcode")
		}
		message = append(message, body...)
		if final {
			if !utf8.Valid(message) {
				return nil, errors.New("invalid Codex control text")
			}
			return message, nil
		}
	}
	return nil, errors.New("codex control frame count exceeded")
}
