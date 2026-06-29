package entity

// USDIDRRate is the latest USD→IDR rate from the exchange-rate provider
// (Open Exchange Rates). Rate is IDR per 1 USD as a decimal string; Timestamp
// is the provider's quote time in epoch seconds.
type USDIDRRate struct {
	Rate      string `json:"rate"`
	Timestamp int64  `json:"timestamp"`
}
