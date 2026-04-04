// Package fintrapay provides a Go client for the FintraPay API.
package fintrapay

// Supported blockchains.
var Chains = []string{"tron", "bsc", "ethereum", "solana", "base", "arbitrum", "polygon"}

// Supported tokens.
var Tokens = []string{"USDT", "USDC", "DAI", "FDUSD", "TUSD", "PYUSD"}

// TokenChains maps each token to its supported blockchains.
var TokenChains = map[string][]string{
	"USDT":  {"tron", "bsc", "ethereum", "solana", "base", "arbitrum", "polygon"},
	"USDC":  {"bsc", "ethereum", "solana", "base", "arbitrum", "polygon"},
	"DAI":   {"bsc", "ethereum", "base", "arbitrum", "polygon"},
	"FDUSD": {"bsc", "ethereum"},
	"TUSD":  {"tron", "bsc", "ethereum"},
	"PYUSD": {"ethereum", "solana"},
}

// Invoice statuses.
const (
	InvoicePending       = "pending"
	InvoiceAwaiting      = "awaiting_selection"
	InvoicePaid          = "paid"
	InvoiceConfirmed     = "confirmed"
	InvoiceExpired       = "expired"
	InvoicePartiallyPaid = "partially_paid"
)

// Payout reasons.
var PayoutReasons = []string{"payment", "refund", "reward", "airdrop", "salary", "other"}

// EarnDurations maps contract duration in months to APY percentage.
var EarnDurations = map[int]float64{
	1:  3.0,
	3:  5.0,
	6:  7.0,
	12: 10.0,
}
