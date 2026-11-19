package metrics

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const fileSourceQuietPeriod = 50 * time.Millisecond

func appendUsageFile(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	_, writeErr := f.WriteString(text)
	closeErr := f.Close()
	require.NoError(t, writeErr)
	require.NoError(t, closeErr)
}

func startFileSourceTest(t *testing.T, path string) (chan ZitiEventMsg, func()) {
	t.Helper()
	src := &fileSource{cfg: &FileSourceConfig{Path: path}}
	// the fixtures produce fewer events than this even with complete-lines off,
	// so an unexpected delivery cannot block the source during cleanup.
	events := make(chan ZitiEventMsg, 8)
	join, err := src.Start(events)
	require.NoError(t, err)
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		done := make(chan struct{})
		go func() {
			src.Stop()
			<-join
			src.t.Cleanup()
			_ = src.ptrF.Close()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("file source did not stop")
		}
	}
	t.Cleanup(stop)
	return events, stop
}

func requireFileLine(t *testing.T, events <-chan ZitiEventMsg, want string) {
	t.Helper()
	select {
	case event := <-events:
		require.Equal(t, want, string(event.Data()))
	case <-time.After(time.Second):
		t.Fatal("file source did not deliver the complete line")
	}
}

func requireNoFileLine(t *testing.T, events <-chan ZitiEventMsg) {
	t.Helper()
	select {
	case event := <-events:
		t.Fatalf("unexpected file line '%s'", event.Data())
	case <-time.After(fileSourceQuietPeriod):
	}
}

func requireFilePointer(t *testing.T, path string, want int64) {
	t.Helper()
	// delivery precedes the pointer write; read through a separate file handle
	// rather than racing the source's seek and write on its pointer handle.
	require.Eventually(t, func() bool {
		data, err := os.ReadFile(path + ".ptr")
		return err == nil && len(data) == 8 && int64(binary.LittleEndian.Uint64(data)) == want
	}, time.Second, time.Millisecond)
}

func TestFileSourceWaitsForNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	const ready = "ready\n"
	const firstHalf = `{"event":`
	require.NoError(t, os.WriteFile(path, []byte(ready+firstHalf), 0o600))
	events, _ := startFileSourceTest(t, path)
	// receiving a complete startup line synchronizes source initialization.
	// the unfinished line is already in the file, so no append race is needed.
	requireFileLine(t, events, "ready")
	requireNoFileLine(t, events)
	appendUsageFile(t, path, "1}\n")
	requireFileLine(t, events, `{"event":1}`)
	requireNoFileLine(t, events)
}

func TestFileSourceKeepsLinesAfterUnfinishedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	const first = "{\"event\":1}\n"
	require.NoError(t, os.WriteFile(path, []byte(first), 0o600))
	events, _ := startFileSourceTest(t, path)
	requireFileLine(t, events, `{"event":1}`)
	appendUsageFile(t, path, `{"event":`)
	requireNoFileLine(t, events)
	requireFilePointer(t, path, int64(len(first)))
	appendUsageFile(t, path, "2}\n{\"event\":3}\n")
	requireFileLine(t, events, `{"event":2}`)
	requireFileLine(t, events, `{"event":3}`)
	requireNoFileLine(t, events)
	requireFilePointer(t, path, int64(len(first)+len("{\"event\":2}\n{\"event\":3}\n")))
}

func TestFileSourceRestartResumesUnfinishedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	const first = "{\"event\":1}\n"
	require.NoError(t, os.WriteFile(path, []byte(first+`{"event":`), 0o600))
	events, stop := startFileSourceTest(t, path)
	requireFileLine(t, events, `{"event":1}`)
	requireNoFileLine(t, events)
	requireFilePointer(t, path, int64(len(first)))
	stop()
	require.Empty(t, events)

	events, _ = startFileSourceTest(t, path)
	appendUsageFile(t, path, "2}\n")
	requireFileLine(t, events, `{"event":2}`)
	requireNoFileLine(t, events)
	requireFilePointer(t, path, int64(len(first)+len("{\"event\":2}\n")))
}
