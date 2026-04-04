# fintrapay-go

Official Go SDK for the [FintraPay](https://fintrapay.io) crypto payment gateway API. Accept stablecoin payments, payment links, subscriptions, deposit API, payouts, withdrawals, and earn yield -- all with automatic HMAC-SHA256 request signing.

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://pkg.go.dev/github.com/Fintra-Ltd/fintrapay-go)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/go-1.21%2B-blue.svg)](https://go.dev/)

---

## Installation

```bash
go get github.com/Fintra-Ltd/fintrapay-go
```

## Quick Start

### Create an Invoice

```go
package main

import (
	"context"
	"fmt"
	"log"

	fintrapay "github.com/Fintra-Ltd/fintrapay-go"
)

func main() {
	client := fintrapay.NewClient("xfp_key_your_api_key", "xfp_secret_your_api_secret")

	// Single-token invoice
	invoice, err := client.CreateInvoice(context.Background(), &fintrapay.CreateInvoiceRequest{
		Amount:     "100.00",
		Currency:   "USDT",
		Blockchain: "tron",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Payment address: %s\n", invoice.PaymentAddress)
	fmt.Printf("Invoice ID: %s\n", invoice.ID)
}
```

### Verify a Webhook

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-FintraPay-Signature")
	if !fintrapay.VerifyWebhookSignature(body, sig, webhookSecret) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var event map[string]interface{}
	json.Unmarshal(body, &event)
	fmt.Printf("Verified event: %s\n", event["type"])
	w.WriteHeader(http.StatusOK)
}
```

## API Reference

All methods accept a `context.Context` as the first parameter and return typed structs. HMAC-SHA256 signing is handled automatically.

### Invoices

| Method | Description |
|--------|-------------|
| `CreateInvoice(ctx, *CreateInvoiceRequest)` | Create a payment invoice |
| `GetInvoice(ctx, invoiceID)` | Get invoice by ID |
| `ListInvoices(ctx, *ListInvoicesParams)` | List invoices with filters |

### Payouts

| Method | Description |
|--------|-------------|
| `CreatePayout(ctx, *CreatePayoutRequest)` | Create a single payout |
| `CreateBatchPayout(ctx, *CreateBatchPayoutRequest)` | Create a batch payout |
| `GetPayout(ctx, payoutID)` | Get payout by ID |
| `ListPayouts(ctx, *ListPayoutsParams)` | List payouts with filters |
| `ListBatchPayouts(ctx, *ListBatchPayoutsParams)` | List batch payouts |
| `GetBatchPayout(ctx, batchID)` | Get batch payout details |

### Withdrawals

| Method | Description |
|--------|-------------|
| `CreateWithdrawal(ctx, *CreateWithdrawalRequest)` | Withdraw to your registered wallet |
| `GetWithdrawal(ctx, withdrawalID)` | Get withdrawal by ID |
| `ListWithdrawals(ctx, *ListWithdrawalsParams)` | List withdrawals |

### Earn

| Method | Description |
|--------|-------------|
| `CreateEarnContract(ctx, *CreateEarnContractRequest)` | Create an Earn contract |
| `GetEarnContract(ctx, contractID)` | Get Earn contract by ID |
| `ListEarnContracts(ctx, *ListEarnContractsParams)` | List Earn contracts |
| `WithdrawEarnInterest(ctx, contractID, amount)` | Withdraw accrued interest (min $10) |
| `BreakEarnContract(ctx, contractID)` | Early-break an Earn contract |
| `GetInterestHistory(ctx, contractID)` | Get daily interest accrual history |

### Refunds

| Method | Description |
|--------|-------------|
| `CreateRefund(ctx, invoiceID, *CreateRefundRequest)` | Create a refund for a paid invoice |
| `GetRefund(ctx, refundID)` | Get refund by ID |
| `ListRefunds(ctx, *ListRefundsParams)` | List all refunds |
| `ListInvoiceRefunds(ctx, invoiceID)` | List refunds for a specific invoice |

### Payment Links

| Method | Description |
|--------|-------------|
| `CreatePaymentLink(ctx, *CreatePaymentLinkRequest)` | Create a reusable payment link |
| `ListPaymentLinks(ctx, *ListPaymentLinksParams)` | List payment links with filters |
| `GetPaymentLink(ctx, linkID)` | Get payment link by ID |
| `UpdatePaymentLink(ctx, linkID, *UpdatePaymentLinkRequest)` | Update a payment link |

### Subscription Plans

| Method | Description |
|--------|-------------|
| `CreateSubscriptionPlan(ctx, *CreateSubscriptionPlanRequest)` | Create a subscription plan |
| `ListSubscriptionPlans(ctx, *ListSubscriptionPlansParams)` | List subscription plans |
| `GetSubscriptionPlan(ctx, planID)` | Get plan by ID |
| `UpdateSubscriptionPlan(ctx, planID, *UpdateSubscriptionPlanRequest)` | Update a subscription plan |

### Subscriptions

| Method | Description |
|--------|-------------|
| `CreateSubscription(ctx, *CreateSubscriptionRequest)` | Create a subscription |
| `ListSubscriptions(ctx, *ListSubscriptionsParams)` | List subscriptions with filters |
| `GetSubscription(ctx, subscriptionID)` | Get subscription with invoice history |
| `CancelSubscription(ctx, subscriptionID, reason)` | Cancel a subscription |
| `PauseSubscription(ctx, subscriptionID)` | Pause an active subscription |
| `ResumeSubscription(ctx, subscriptionID)` | Resume a paused subscription |

### Deposit API

| Method | Description |
|--------|-------------|
| `CreateDepositUser(ctx, *CreateDepositUserRequest)` | Register end user for deposits |
| `GetDepositUser(ctx, externalUserID)` | Get user with addresses and balances |
| `ListDepositUsers(ctx, *ListDepositUsersParams)` | List deposit users |
| `UpdateDepositUser(ctx, externalUserID, *UpdateDepositUserRequest)` | Update user |
| `CreateDepositAddress(ctx, externalUserID, blockchain)` | Generate address for a chain |
| `CreateAllDepositAddresses(ctx, externalUserID)` | Generate addresses for all 7 chains |
| `ListDepositAddresses(ctx, externalUserID)` | List all addresses for a user |
| `ListDeposits(ctx, *ListDepositsParams)` | List deposit events |
| `GetDeposit(ctx, depositID)` | Get single deposit detail |
| `ListDepositBalances(ctx, externalUserID)` | Get per-token per-chain balances |

### Balance & Fees

| Method | Description |
|--------|-------------|
| `GetBalance(ctx)` | Get custodial balances across all chains |
| `EstimateFees(ctx, *EstimateFeesRequest)` | Estimate transaction fees |

### Support Tickets

| Method | Description |
|--------|-------------|
| `CreateTicket(ctx, *CreateTicketRequest)` | Create a support ticket |
| `ListTickets(ctx, *ListTicketsParams)` | List support tickets |
| `GetTicket(ctx, ticketID)` | Get ticket by ID |
| `ReplyTicket(ctx, ticketID, message)` | Reply to a support ticket |

## Error Handling

The SDK returns typed errors for different scenarios:

```go
import (
	"errors"
	"fmt"

	fintrapay "github.com/Fintra-Ltd/fintrapay-go"
)

invoice, err := client.CreateInvoice(ctx, &fintrapay.CreateInvoiceRequest{
	Amount:     "100.00",
	Currency:   "USDT",
	Blockchain: "tron",
})
if err != nil {
	var authErr *fintrapay.AuthError
	var valErr *fintrapay.ValidationError
	var rateErr *fintrapay.RateLimitError
	var apiErr *fintrapay.APIError

	switch {
	case errors.As(err, &authErr):
		// Invalid API key or secret (HTTP 401)
		fmt.Printf("Auth failed: %s\n", authErr.Message)
	case errors.As(err, &valErr):
		// Invalid request parameters (HTTP 422)
		fmt.Printf("Validation error: %s\n", valErr.Message)
	case errors.As(err, &rateErr):
		// Too many requests (HTTP 429)
		fmt.Printf("Rate limited. Retry after %d seconds\n", rateErr.RetryAfter)
	case errors.As(err, &apiErr):
		// Any other API error
		fmt.Printf("API error (%d): %s\n", apiErr.StatusCode, apiErr.Message)
	default:
		// Network or other error
		fmt.Printf("Error: %v\n", err)
	}
}
```

## Webhook Verification

Always verify webhook signatures before processing events. Use the raw request body -- do NOT parse JSON first.

### net/http

```go
import (
	"encoding/json"
	"io"
	"net/http"

	fintrapay "github.com/Fintra-Ltd/fintrapay-go"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-FintraPay-Signature")
	if !fintrapay.VerifyWebhookSignature(body, sig, webhookSecret) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var event map[string]interface{}
	json.Unmarshal(body, &event)
	// process event...

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
```

## Configuration Options

```go
// Custom base URL
client := fintrapay.NewClient(apiKey, apiSecret, fintrapay.WithBaseURL("https://custom.api.url/v1"))

// Custom timeout
client := fintrapay.NewClient(apiKey, apiSecret, fintrapay.WithTimeout(60 * time.Second))

// Custom HTTP client
client := fintrapay.NewClient(apiKey, apiSecret, fintrapay.WithHTTPClient(myHTTPClient))
```

## Requirements

- Go 1.21 or later
- Zero external dependencies (uses Go standard library only)

## Supported Chains & Tokens

7 blockchains: TRON, BSC, Ethereum, Solana, Base, Arbitrum, Polygon

6 stablecoins: USDT, USDC, DAI, FDUSD, TUSD, PYUSD

## Links

- [FintraPay Homepage](https://fintrapay.io)
- [API Documentation](https://fintrapay.io/docs)
- [GitHub Repository](https://github.com/Fintra-Ltd/fintrapay-go)
- [Go Package Reference](https://pkg.go.dev/github.com/Fintra-Ltd/fintrapay-go)

## License

MIT License. See [LICENSE](LICENSE) for details.
