package subordinate

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageFraming(t *testing.T) {
	h := NewMessageHandler()
	var boots int
	var received []Message
	h.BootHandler = func(kind string, msg Message) {
		require.Equal(t, BootMessage, kind)
		require.Equal(t, "token", msg["token"])
		boots++
	}
	h.MessageHandler = func(msg Message) { received = append(received, msg) }
	for _, chunk := range []string{" \n\n{\"msg\":\"bo", "ot\",\"token\":\"to", "ken\"}\n{\"msg\":\"access\",\"path\":\"first\"}\nplain", " text\n{\"msg\":\"access\",\"path\":\"last\"}\n"} {
		h.Tail([]byte(chunk))
	}
	h.Close()
	h.Close()
	require.NoError(t, h.Err())
	require.Equal(t, 1, boots)
	require.Len(t, received, 3)
	require.Equal(t, "first", received[0]["path"])
	require.Equal(t, "plain text", received[1][RawMessage])
	require.Equal(t, "last", received[2]["path"])
	select {
	case <-h.BootComplete:
	default:
		t.Fatal("boot was not completed")
	}
}

func TestEOFAndOverflowFailStartup(t *testing.T) {
	for _, tc := range []struct {
		name     string
		data     []byte
		overflow bool
	}{
		{name: "empty"}, {name: "partial", data: []byte(`{"msg":"boot","token":"token"}`)},
		{name: "overflow", data: bytes.Repeat([]byte("x"), MaxLineSize+1), overflow: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewMessageHandler()
			var kind string
			h.BootHandler = func(k string, _ Message) { kind = k }
			h.Tail(tc.data)
			h.Close()
			require.Equal(t, ErrorMessage, kind)
			if tc.overflow {
				require.ErrorContains(t, h.Err(), "exceeds")
			} else {
				require.ErrorIs(t, h.Err(), io.ErrUnexpectedEOF)
			}
			select {
			case <-h.BootComplete:
			default:
				t.Fatal("startup did not fail")
			}
		})
	}
}

func TestProtocolLimitAndMalformedMessages(t *testing.T) {
	h := NewMessageHandler()
	var messages, malformed int
	h.MessageHandler = func(Message) { messages++ }
	h.MalformedHandler = func(Message) { malformed++ }
	h.Tail(bytes.Repeat([]byte("x"), MaxLineSize))
	h.Tail([]byte("\n{\"msg\":7}\n{}\n{\"msg\":\"boot\",\"token\":\"t\"}\n"))
	require.NoError(t, h.Err())
	require.Equal(t, 1, messages)
	require.Equal(t, 2, malformed)
	h.Tail(bytes.Repeat([]byte("x"), MaxLineSize+1))
	require.ErrorContains(t, h.Err(), "exceeds")
	require.Equal(t, 2, messages)
}

func TestEmptyErrorMessageCannotBootSuccessfully(t *testing.T) {
	h := NewMessageHandler()
	var msg Message
	h.BootHandler = func(_ string, in Message) { msg = in }
	h.Tail([]byte("{\"msg\":\"error\"}\n"))
	require.Error(t, h.Err())
	require.NotEmpty(t, msg[ErrorMessage])
}
