package session

import (
	"testing"
	"time"

	"github.com/mtgo-labs/mtgo/tg"
)

func TestFloodWaitHandling(t *testing.T) {
	wait, ok := parseFloodWait("FLOOD_WAIT_5")
	if !ok {
		t.Fatal("parseFloodWait() ok = false, want true")
	}
	if wait != 5*time.Second {
		t.Fatalf("parseFloodWait() = %v, want 5s", wait)
	}
	if _, ok := parseFloodWait("PHONE_CODE_INVALID"); ok {
		t.Fatal("parseFloodWait(non-flood) ok = true, want false")
	}
}

func TestFloodWaitQueue(t *testing.T) {
	q := &FloodWaitQueue{}
	query := &tg.PingRequest{PingID: 1}
	q.Delay(query, 10, -time.Millisecond)

	ready := q.Ready()
	if len(ready) != 1 {
		t.Fatalf("Ready() len = %d, want 1", len(ready))
	}
	if ready[0].Query != query || ready[0].MsgID != 10 {
		t.Fatalf("Ready()[0] = %#v, want query/msg_id", ready[0])
	}
	if ready := q.Ready(); len(ready) != 0 {
		t.Fatalf("Ready() after drain len = %d, want 0", len(ready))
	}
	q.Delay(query, 11, time.Hour)
	q.Cleanup()
	if ready := q.Ready(); len(ready) != 0 {
		t.Fatalf("Ready() after cleanup len = %d, want 0", len(ready))
	}
}
