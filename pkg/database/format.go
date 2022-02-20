package database

type Balance struct {
	Free  float64 `json:"free"`
	Used  float64 `json:"used"`
	Total float64 `json:"total"`
}

type Fee struct {
	Maker float64 `json:"maker"`
	Taker float64 `json:"taker"`
}

type Order struct {
	Id          string  `json:"order_id"`
	Type        string  `json:"type"`
	Side        string  `json:"side"`
	CreateTime  string  `json:"create_time_ms"`
	UpdateTime  string  `json:"update_time_ms"`
	Price       float64 `json:"price"`
	Amount      float64 `json:"amount"`
	LeftAmount  float64 `json:"left"`
	Status      string  `json:"status"`
	Fee         float64 `json:"fee"`
	FeeCurrency string  `json:"fee_currency"`
}

type OrderBook struct {
	Asks [][2]float64 `json:"asks"`
	Bids [][2]float64 `json:"bids"`
}
