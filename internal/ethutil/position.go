package ethutil

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

var v2ReservedBits = new(big.Int).Lsh(new(big.Int).SetUint64(1<<64-1), 40)

// IsV2PositionID reports whether a decimal asset ID is a Protocol V2 position ID.
// V2 IDs keep bits 40 to 103 clear, CTF token IDs are hashes and never do.
func IsV2PositionID(id string) bool {
	n, ok := new(big.Int).SetString(strings.TrimSpace(id), 10)
	if !ok || n.Sign() < 0 || n.BitLen() > 256 {
		return false
	}
	return new(big.Int).And(n, v2ReservedBits).Sign() == 0
}

// V2ConditionWord ABI encodes a V2 condition ID as bytes31. It accepts the
// bytes31 form and the 32 byte form the APIs return, which ends in 00 or 01.
func V2ConditionWord(conditionID string) ([]byte, error) {
	raw, err := hex.DecodeString(StripHexPrefix(strings.TrimSpace(conditionID)))
	if err != nil {
		return nil, fmt.Errorf("v2 condition id %q: %w", conditionID, err)
	}
	switch {
	case len(raw) == 31:
	case len(raw) == 32 && raw[31] <= 1:
		raw = raw[:31]
	default:
		return nil, fmt.Errorf("v2 condition id %q must be bytes31, or bytes32 ending in 00 or 01", conditionID)
	}
	return append(raw, 0), nil
}

// V2PositionID returns the decimal position ID of an outcome of a V2 condition.
func V2PositionID(conditionWord []byte, outcomeIndex int) string {
	id := new(big.Int).SetBytes(conditionWord)
	return id.Or(id, big.NewInt(int64(outcomeIndex))).String()
}
