package exchange

import (
	"Markets/pkg/database"
	"reflect"
	"testing"
)

func TestUpdateOrderBook(t *testing.T) {
	ob := &database.OrderBook{
		Asks: map[string]string{
			"0.0000026400": "1000000",
			"0.0000026500": "20000",
		},
		Bids: map[string]string{
			"0.0000026200": "1000000",
			"0.0000026000": "20000",
			"0.0000025000": "10000",
		},
	}

	asksData := [][]string{
		{"0.0000012345", "19929"},
		{"0.0000023456", "45644"},
	}

	bidsData := [][]string{
		{"0.0000026000", "0"},
		{"0.0000034567", "78978"},
	}

	updateOrderBook(false, ob, asksData, bidsData)

	if !reflect.DeepEqual(ob, &database.OrderBook{
		Asks: map[string]string{
			"0.0000026400": "1000000",
			"0.0000026500": "20000",
			"0.0000012345": "19929",
			"0.0000023456": "45644",
		},
		Bids: map[string]string{
			"0.0000026200": "1000000",
			"0.0000025000": "10000",
			"0.0000034567": "78978",
		},
	}){
		t.Error("OrderBook not incremental updated correctly")
	}


	asksData = [][]string{
		{"0.0000012345", "19929"},
		{"0.0000023456", "45644"},
	}

	bidsData = [][]string{
		{"0.0000026000", "12345"},
		{"0.0000034567", "78978"},
	}

	updateOrderBook(true, ob, asksData, bidsData)

	if !reflect.DeepEqual(ob, &database.OrderBook{
		Asks: map[string]string{
			"0.0000012345": "19929",
			"0.0000023456": "45644",
		},
		Bids: map[string]string{
			"0.0000026000": "12345",
			"0.0000034567": "78978",
		},
	}){
		t.Error("OrderBook not fully updated correctly")
	}
}
