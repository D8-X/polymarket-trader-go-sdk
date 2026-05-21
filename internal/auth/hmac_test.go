package auth

import "testing"

func TestSignMessageGolden(t *testing.T) {
	const (
		secret = "06iZVHeK0RaXlk1dMfx35xeVYLuw3F1v9XT6RyoWFfQ="
		ts     = "1716000000"
		method = "GET"
		path   = "/balance-allowance"
	)
	got, err := SignMessage(secret, ts, method, path, nil)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	const want = "8SCLVcNKOLeZLDUA-uXdciogvKdj1DOQaTYgL1xLvsE="
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestSignMessageGoldenWithBody(t *testing.T) {
	const (
		secret = "06iZVHeK0RaXlk1dMfx35xeVYLuw3F1v9XT6RyoWFfQ="
		ts     = "1716000000"
		method = "POST"
		path   = "/order"
	)
	body := []byte(`{"orderID":"abc"}`)
	got, err := SignMessage(secret, ts, method, path, body)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	const want = "0oSK_TBcfF8NWeDGPzPNSU3m0BbknYFBt6xtrccKlkM="
	if got != want {
		t.Errorf("sig mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}
