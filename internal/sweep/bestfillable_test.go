package sweep

import (
	"testing"

	"github.com/D8-X/polymarket-trader-go-sdk/v2/internal/models"
)

func TestBestFillablePriceBuy(t *testing.T) {
	book := &models.OrderBook{
		Asks: []models.OrderBookLevel{
			{Price: "0.50", Size: "10"},
			{Price: "0.52", Size: "10"},
			{Price: "0.55", Size: "10"},
		},
	}
	p, err := BestFillablePrice(book, Buy, 15)
	if err != nil {
		t.Fatal(err)
	}
	if p != 0.52 {
		t.Errorf("got %v want 0.52", p)
	}
}

func TestBestFillablePriceSell(t *testing.T) {
	book := &models.OrderBook{
		Bids: []models.OrderBookLevel{
			{Price: "0.55", Size: "10"},
			{Price: "0.52", Size: "10"},
		},
	}
	p, err := BestFillablePrice(book, Sell, 5)
	if err != nil {
		t.Fatal(err)
	}
	if p != 0.55 {
		t.Errorf("got %v want 0.55", p)
	}
}

func TestBestFillablePriceErrsWhenThin(t *testing.T) {
	book := &models.OrderBook{
		Asks: []models.OrderBookLevel{{Price: "0.50", Size: "5"}},
	}
	if _, err := BestFillablePrice(book, Buy, 100); err == nil {
		t.Error("expected error when book is too thin")
	}
}

func TestBestFillablePriceErrsOnBadInputs(t *testing.T) {
	if _, err := BestFillablePrice(nil, Buy, 10); err == nil {
		t.Error("expected nil-book error")
	}
	book := &models.OrderBook{Asks: []models.OrderBookLevel{{Price: "0.50", Size: "5"}}}
	if _, err := BestFillablePrice(book, Buy, 0); err == nil {
		t.Error("expected zero-size error")
	}
}
