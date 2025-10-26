package models

type Price struct {
	CurrencyPair string  `json:"currency_pair"`
	Price        float64 `json:"price"`
	Timestamp    int64   `json:"timestamp"`
}