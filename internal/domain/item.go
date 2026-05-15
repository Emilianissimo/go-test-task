package domain

import (
	"github.com/shopspring/decimal"
)

type ItemResponse struct {
	MarketHashName string           `json:"market_hash_name"`
	Currency       string           `json:"currency"`
	MinPrice       *decimal.Decimal `json:"min_price"`
}

type MergedItem struct {
	MarketHashName      string           `json:"market_hash_name"`
	Currency            string           `json:"currency"`
	MinPriceTradable    *decimal.Decimal `json:"tradable_min_price"`
	MinPriceNonTradable *decimal.Decimal `json:"non_tradable_min_price"`
}
