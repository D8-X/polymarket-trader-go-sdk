package polytrade

import "testing"

func TestPrepareCloseProducesSellOrder(t *testing.T) {
	ob := NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
	signed, err := ob.PrepareClose("100", 0.55, 10, "k", ClosePositionOpts{})
	if err != nil {
		t.Fatalf("prepare close: %v", err)
	}
	if signed.Order.Side != SELL {
		t.Errorf("side: got %s want %s", signed.Order.Side, SELL)
	}
	if signed.OrderType != OrderTypeFOK {
		t.Errorf("order type: got %s want default %s", signed.OrderType, OrderTypeFOK)
	}
}

func TestPrepareCloseHonoursOrderTypeOverride(t *testing.T) {
	ob := NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
	signed, err := ob.PrepareClose("100", 0.55, 10, "k", ClosePositionOpts{OrderType: OrderTypeGTC})
	if err != nil {
		t.Fatalf("prepare close: %v", err)
	}
	if signed.OrderType != OrderTypeGTC {
		t.Errorf("order type: got %s want %s", signed.OrderType, OrderTypeGTC)
	}
	if signed.Order.Side != SELL {
		t.Errorf("side: got %s want %s", signed.Order.Side, SELL)
	}
}

func TestPrepareCloseAmounts(t *testing.T) {
	ob := NewOrderBuilder(testDepositWallet, CTFExchange, testPrivateKey, SignatureTypePoly1271)
	signed, err := ob.PrepareClose("100", 0.55, 10, "k", ClosePositionOpts{TickSize: "0.01"})
	if err != nil {
		t.Fatalf("prepare close: %v", err)
	}
	if signed.Order.MakerAmount != "10000000" {
		t.Errorf("maker amount: got %s want 10000000", signed.Order.MakerAmount)
	}
	if signed.Order.TakerAmount != "5500000" {
		t.Errorf("taker amount: got %s want 5500000", signed.Order.TakerAmount)
	}
}
