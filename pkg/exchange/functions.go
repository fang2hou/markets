package exchange

import "Markets/pkg/database"

func updateOrderBook(
	fullMode bool,
	originalOrderBook *database.OrderBook,
	asksData [][]string,
	bidsData [][]string,
) {
	if fullMode {
		originalOrderBook.Asks = make(map[string]string)
		originalOrderBook.Bids = make(map[string]string)
	}

	for _, ask := range asksData {
		if !fullMode && ask[1] == "0" {
			delete(originalOrderBook.Asks, ask[0])
		} else {
			originalOrderBook.Asks[ask[0]] = ask[1]
		}
	}

	for _, bid := range bidsData {
		if !fullMode && bid[1] == "0" {
			delete(originalOrderBook.Bids, bid[0])
		} else {
			originalOrderBook.Bids[bid[0]] = bid[1]
		}
	}
}
