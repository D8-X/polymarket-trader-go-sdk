package clob

import (
	"math/big"
	"strings"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/consts"
	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func fakRemainingSize(signed *models.SignedOrder, resp *models.PlaceOrderResponse) string {
	if signed == nil || resp == nil || signed.OrderType != consts.OrderTypeFAK {
		return ""
	}
	if resp.Status == consts.OrderStatusDelayed {
		return ""
	}
	orderedBase, filled := signed.Order.TakerAmount, resp.TakingAmount
	if signed.Order.Side == consts.SELL {
		orderedBase, filled = signed.Order.MakerAmount, resp.MakingAmount
	}
	ordered, ok := new(big.Rat).SetString(orderedBase)
	if !ok {
		return ""
	}
	ordered.Quo(ordered, new(big.Rat).SetInt64(int64(consts.AmountScale)))
	f := new(big.Rat)
	if filled != "" {
		if _, ok := f.SetString(filled); !ok {
			return ""
		}
	}
	rem := ordered.Sub(ordered, f)
	if rem.Sign() < 0 {
		return "0"
	}
	s := rem.FloatString(6)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}
