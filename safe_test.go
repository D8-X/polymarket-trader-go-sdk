package polytrade

import "testing"

func TestDeriveSafeAddressGolden(t *testing.T) {
	eoa := testEOA(t)
	got := DeriveSafeAddress(eoa)
	const want = "0x5311d68AC78597E9DDE563Ce19DE4C998779B83c"
	if got != want {
		t.Errorf("safe mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

func TestDerivePolymarketProxyGolden(t *testing.T) {
	eoa := testEOA(t)
	got := DerivePolymarketProxy(eoa)
	const want = "0xc36e0875e694631E4bc343eEA84b7b404e335441"
	if got != want {
		t.Errorf("proxy mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}
