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

func RoundTo10k(v int64) int64 {
	return (v / 10000) * 10000
}
