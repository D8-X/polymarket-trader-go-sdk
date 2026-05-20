package ethutil

import "math/big"

func StripHexPrefix(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}

func ParseBigInt(s string) []byte {
	n := new(big.Int)
	if len(s) > 2 && s[:2] == "0x" {
		n.SetString(s[2:], 16)
	} else {
		n.SetString(s, 10)
	}
	return n.Bytes()
}

func ParseAddress(s string) []byte {
	n := new(big.Int)
	if len(s) > 2 {
		n.SetString(s[2:], 16)
	}
	return n.Bytes()
}

// ParseBytes32 returns a left-padded 32-byte representation of a hex string.
func ParseBytes32(s string) []byte {
	out := make([]byte, 32)
	if len(s) == 0 {
		return out
	}
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	n := new(big.Int)
	n.SetString(s, 16)
	b := n.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(out[32-len(b):], b)
	return out
}

