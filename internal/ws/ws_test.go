package ws

import "testing"

func TestParseEventsHandlesBatch(t *testing.T) {
	msg := []byte(`[{"event_type":"a"},{"event_type":"b"}]`)
	out := ParseEvents(msg)
	if len(out) != 2 || out[0].Type != "a" || out[1].Type != "b" {
		t.Errorf("got %+v", out)
	}
}
