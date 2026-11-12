package codex

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func serverControlFrame(opcode byte, final bool, body []byte) []byte {
	if final {
		opcode |= 0x80
	}
	frame := []byte{opcode}
	switch {
	case len(body) < 126:
		frame = append(frame, byte(len(body)))
	case len(body) <= 65535:
		frame = binary.BigEndian.AppendUint16(append(frame, 126), uint16(len(body)))
	default:
		frame = binary.BigEndian.AppendUint64(append(frame, 127), uint64(len(body)))
	}
	return append(frame, body...)
}

func TestSharedControlWireFragmentedTextAndPing(t *testing.T) {
	frames := serverControlFrame(1, false, []byte(`{"res`))
	frames = append(frames, serverControlFrame(9, true, []byte("ping"))...)
	frames = append(frames, serverControlFrame(0, true, []byte(`ult":{}}`))...)
	var sent bytes.Buffer
	wire := sharedOwnerWire{r: bufio.NewReader(bytes.NewReader(frames)), w: &sent}
	got, err := wire.read()
	if err != nil || string(got) != `{"result":{}}` {
		t.Fatal(string(got), err)
	}
	if sent.Bytes()[0] != 0x8a || sent.Bytes()[1] != 0x84 {
		t.Fatal("ping was not answered with masked pong")
	}
}

func TestSharedControlWireMasksWritesAndReadsSizeBoundaries(t *testing.T) {
	for _, size := range []int{0, 125, 126, 65535, 65536} {
		body := bytes.Repeat([]byte("x"), size)
		var sent bytes.Buffer
		wire := sharedOwnerWire{r: bufio.NewReader(bytes.NewReader(serverControlFrame(1, true, body))), w: &sent}
		if err := wire.write(1, body); err != nil {
			t.Fatal(err)
		}
		frame := sent.Bytes()
		if frame[1]&0x80 == 0 {
			t.Fatal("unmasked client frame")
		}
		offset := 2
		if frame[1]&0x7f == 126 {
			offset += 2
		}
		if frame[1]&0x7f == 127 {
			offset += 8
		}
		for i, b := range frame[offset+4:] {
			if b^frame[offset+i%4] != body[i] {
				t.Fatal("invalid mask")
			}
		}
		got, err := wire.read()
		if err != nil || !bytes.Equal(got, body) {
			t.Fatal(size, err)
		}
	}
}

func TestSharedControlWireRejectsMalformedAndOversizedFrames(t *testing.T) {
	tooLarge := binary.BigEndian.AppendUint64([]byte{0x81, 127}, sharedWireLimit+1)
	for _, frame := range [][]byte{{0xc1, 0}, {0x81, 0x80}, {0x80, 0}, {0x82, 0}, {0x09, 0}, {0x89, 126, 0, 126}, {0x81, 1, 0xff}, tooLarge, {0x88, 0}, {0x81, 5, 'x'}} {
		wire := sharedOwnerWire{r: bufio.NewReader(bytes.NewReader(frame)), w: io.Discard}
		if _, err := wire.read(); err == nil {
			t.Fatalf("accepted bad frame: %x", frame)
		}
	}
}

func TestSharedControlUpgradeValidatesHandshake(t *testing.T) {
	for _, valid := range []bool{true, false} {
		client, server := net.Pipe()
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer server.Close()
			req, err := http.ReadRequest(bufio.NewReader(server))
			if err != nil {
				return
			}
			digest := sha1.Sum([]byte(req.Header.Get("Sec-WebSocket-Key") + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
			accept := base64.StdEncoding.EncodeToString(digest[:])
			if !valid {
				accept = "invalid"
			}
			_, _ = io.WriteString(server, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: keep-alive, Upgrade\r\nSec-WebSocket-Accept: "+accept+"\r\n\r\n")
		}()
		_, err := upgradeSharedOwner(client, client)
		_ = client.Close()
		<-done
		if (err == nil) != valid {
			t.Fatal(valid, err)
		}
	}
}

func TestSharedControlRPCIgnoresOtherResponsesAndSanitizesErrors(t *testing.T) {
	for _, tc := range []struct {
		message string
		want    error
	}{
		{"thread already has an active or pending turn", errSharedOwnerBusy},
		{"queued submission not found: mine", errSharedQueueConsumed},
		{"private-auth-token", nil},
	} {
		body, _ := json.Marshal(map[string]any{"id": 1, "error": map[string]any{"message": tc.message}})
		input := serverControlFrame(1, true, []byte(`{"id":1,"method":"item/tool/requestUserInput","params":{}}`))
		input = append(input, serverControlFrame(1, true, []byte(`{"id":900,"result":{}}`))...)
		input = append(input, serverControlFrame(1, true, body)...)
		client := sharedOwnerClient{wire: &sharedOwnerWire{r: bufio.NewReader(bytes.NewReader(input)), w: io.Discard}}
		_, err := client.request("thread/queue/start", map[string]any{"threadId": "native", "queuedSubmissionId": "mine"})
		if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || strings.Contains(err.Error(), "private-auth-token") {
			t.Fatal(err)
		}
	}
}
