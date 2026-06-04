package clob

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func fakRemainingSize(signed *models.SignedOrder, resp *models.PlaceOrderResponse) (string, error) {
	if signed == nil || resp == nil || signed.OrderType != consts.OrderTypeFAK || resp.Status == consts.OrderStatusDelayed {
		return "", nil
	}
	orderedBase, filled := signed.Order.TakerAmount, resp.TakingAmount
	if signed.Order.Side == consts.SELL {
		orderedBase, filled = signed.Order.MakerAmount, resp.MakingAmount
	}
	ordered, ok := new(big.Rat).SetString(orderedBase)
	if !ok {
		return "", fmt.Errorf("fak remaining size: invalid ordered amount %q", orderedBase)
	}
	ordered.Quo(ordered, new(big.Rat).SetInt64(int64(consts.AmountScale)))
	f := new(big.Rat)
	if filled != "" {
		if _, ok := f.SetString(filled); !ok {
			return "", fmt.Errorf("fak remaining size: invalid filled amount %q", filled)
		}
	}
	rem := ordered.Sub(ordered, f)
	if rem.Sign() < 0 {
		return "0", nil
	}
	s := rem.FloatString(6)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s, nil
}
