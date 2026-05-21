package polytrade

import "github.com/D8-X/polymarket-trader-go-sdk/v2/internal/sign"

func (ob *OrderBuilder) signOrder(order OrderFields) (string, error) {
	return sign.Order(ob.privateKeyHex, ob.ctfExchangeAddress, order)
}
