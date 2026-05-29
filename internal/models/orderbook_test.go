package models

import "testing"

func TestOrderBookBestBidAskMid(t *testing.T) {
	book := OrderBook{
		Bids: []OrderBookLevel{
			{Price: "0.01", Size: "100"},
			{Price: "0.48", Size: "2.47"},
			{Price: "0.20", Size: "50"},
		},
		Asks: []OrderBookLevel{
			{Price: "0.99", Size: "100"},
			{Price: "0.49", Size: "435.75"},
			{Price: "0.75", Size: "50"},
		},
	}
	bp, bs, bok := book.BestBid()
	if !bok || bp != 0.48 || bs != 2.47 {
		t.Errorf("BestBid: got %v/%v ok=%v want 0.48/2.47 ok=true", bp, bs, bok)
	}
	ap, as, aok := book.BestAsk()
	if !aok || ap != 0.49 || as != 435.75 {
		t.Errorf("BestAsk: got %v/%v ok=%v want 0.49/435.75 ok=true", ap, as, aok)
	}
	mid, ok := book.Mid()
	if !ok || mid != 0.485 {
		t.Errorf("Mid: got %v ok=%v want 0.485 ok=true", mid, ok)
	}
}

func TestOrderBookBestEmpty(t *testing.T) {
	book := OrderBook{}
	if _, _, ok := book.BestBid(); ok {
		t.Error("BestBid on empty book should not be ok")
	}
	if _, _, ok := book.BestAsk(); ok {
		t.Error("BestAsk on empty book should not be ok")
	}
	if _, ok := book.Mid(); ok {
		t.Error("Mid on empty book should not be ok")
	}
}

func TestOrderBookMidOnlyOneSide(t *testing.T) {
	book := OrderBook{
		Bids: []OrderBookLevel{{Price: "0.50", Size: "10"}},
	}
	if _, ok := book.Mid(); ok {
		t.Error("Mid should be unavailable when only one side has levels")
	}
}
