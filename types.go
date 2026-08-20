package fintrapay

import "encoding/json"

// ── Invoice types ──────────────────────────────────────────────────

// CreateInvoiceRequest is the request body for creating an invoice.
type CreateInvoiceRequest struct {
	Amount         string   `json:"amount"`
	Currency       string   `json:"currency,omitempty"`
	Blockchain     string   `json:"blockchain,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	AcceptedTokens []string `json:"accepted_tokens,omitempty"`
	AcceptedChains []string `json:"accepted_chains,omitempty"`
	ExternalID     string   `json:"external_id,omitempty"`
	ExpiryMinutes  int      `json:"expiry_minutes,omitempty"`
	ExpiresAt      string   `json:"expires_at,omitempty"`
	SuccessURL     string   `json:"success_url,omitempty"`
	CancelURL      string   `json:"cancel_url,omitempty"`
}

// Invoice represents a payment invoice.
type Invoice struct {
	ID             string          `json:"id"`
	MerchantID     string          `json:"merchant_id"`
	ExternalID     string          `json:"external_id,omitempty"`
	Mode           string          `json:"mode"`
	Amount         string          `json:"amount"`
	Currency       string          `json:"currency"`
	Blockchain     string          `json:"blockchain"`
	Network        string          `json:"network"`
	PaymentAddress string          `json:"payment_address"`
	CheckoutURL    string          `json:"checkout_url,omitempty"`
	Status         string          `json:"status"`
	PaymentTxHash  string          `json:"payment_tx_hash,omitempty"`
	ReceivedAmount string          `json:"received_amount,omitempty"`
	Confirmations  int             `json:"confirmations,omitempty"`
	FeeAmount      string          `json:"fee_amount,omitempty"`
	WebhookStatus  string          `json:"webhook_status,omitempty"`
	PaidAt         string          `json:"paid_at,omitempty"`
	ExpiresAt      string          `json:"expires_at,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
	Extra          json.RawMessage `json:"-"`
}

// ListInvoicesParams holds query parameters for listing invoices.
type ListInvoicesParams struct {
	Status     string `json:"status,omitempty"`
	Blockchain string `json:"blockchain,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Mode       string `json:"mode,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}

// InvoiceList is the response for listing invoices.
type InvoiceList struct {
	Invoices   []Invoice `json:"invoices"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalCount int       `json:"total_count"`
}

// ── Payout types ───────────────────────────────────────────────────

// CreatePayoutRequest is the request body for creating a single payout.
type CreatePayoutRequest struct {
	ToAddress  string `json:"to_address"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Blockchain string `json:"blockchain"`
	Reason     string `json:"reason,omitempty"`
	Reference  string `json:"reference,omitempty"`
	// FeeDeduction is "from_amount" (default) — the recipient gets amount
	// minus fees — or "from_balance", where the recipient gets exactly
	// amount and the fees are debited on top. Empty means from_amount.
	FeeDeduction string `json:"fee_deduction,omitempty"`
}

// Payout represents a single payout.
type Payout struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	BatchID     string `json:"batch_id,omitempty"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Blockchain  string `json:"blockchain"`
	Network     string `json:"network"`
	ToAddress   string `json:"to_address"`
	Reason      string `json:"reason"`
	Reference   string `json:"reference,omitempty"`
	Status      string `json:"status"`
	AdminNotes  string `json:"admin_notes,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"`
	GasFee      string `json:"gas_fee,omitempty"`
	CreatedAt   string `json:"created_at"`
	ApprovedAt  string `json:"approved_at,omitempty"`
	BroadcastAt string `json:"broadcast_at,omitempty"`
	ConfirmedAt string `json:"confirmed_at,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// BatchRecipient is a single recipient in a batch payout.
type BatchRecipient struct {
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
	Reference string `json:"reference,omitempty"`
}

// CreateBatchPayoutRequest is the request body for creating a batch payout.
type CreateBatchPayoutRequest struct {
	Currency   string           `json:"currency"`
	Blockchain string           `json:"blockchain"`
	Recipients []BatchRecipient `json:"recipients"`
}

// BatchPayout represents a batch payout response.
type BatchPayout struct {
	ID          string   `json:"id"`
	MerchantID  string   `json:"merchant_id"`
	TotalAmount string   `json:"total_amount"`
	TotalCount  int      `json:"total_count"`
	Currency    string   `json:"currency"`
	Blockchain  string   `json:"blockchain"`
	Network     string   `json:"network"`
	Status      string   `json:"status"`
	Payouts     []Payout `json:"payouts,omitempty"`
	CreatedAt   string   `json:"created_at"`
	ApprovedAt  string   `json:"approved_at,omitempty"`
	CompletedAt string   `json:"completed_at,omitempty"`
	UpdatedAt   string   `json:"updated_at"`
}

// ListPayoutsParams holds query parameters for listing payouts.
type ListPayoutsParams struct {
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// PayoutList is the response for listing payouts.
type PayoutList struct {
	Payouts    []Payout `json:"payouts"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalCount int      `json:"total_count"`
}

// ── Withdrawal types ───────────────────────────────────────────────

// CreateWithdrawalRequest is the request body for creating a withdrawal.
type CreateWithdrawalRequest struct {
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Blockchain string `json:"blockchain"`
	// ToAddress is optional; when empty the withdrawal goes to the wallet
	// registered for this chain on your merchant profile.
	ToAddress string `json:"to_address,omitempty"`
	// FeeDeduction is "from_amount" (default) — the recipient gets amount
	// minus fees — or "from_balance", where the recipient gets exactly
	// amount and the fees are debited on top. Empty means from_amount.
	FeeDeduction string `json:"fee_deduction,omitempty"`
}

// Withdrawal represents a withdrawal.
type Withdrawal struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Blockchain  string `json:"blockchain"`
	Network     string `json:"network"`
	ToAddress   string `json:"to_address"`
	Status      string `json:"status"`
	TxHash      string `json:"tx_hash,omitempty"`
	GasFee      string `json:"gas_fee,omitempty"`
	CreatedAt   string `json:"created_at"`
	ApprovedAt  string `json:"approved_at,omitempty"`
	BroadcastAt string `json:"broadcast_at,omitempty"`
	ConfirmedAt string `json:"confirmed_at,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// ListWithdrawalsParams holds query parameters for listing withdrawals.
type ListWithdrawalsParams struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// WithdrawalList is the response for listing withdrawals.
type WithdrawalList struct {
	Withdrawals []Withdrawal `json:"withdrawals"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
	TotalCount  int          `json:"total_count"`
}

// ── Earn types ─────────────────────────────────────────────────────

// CreateEarnContractRequest is the request body for creating an Earn contract.
type CreateEarnContractRequest struct {
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	Blockchain     string `json:"blockchain"`
	DurationMonths int    `json:"duration_months"`
}

// EarnContract represents an Earn contract.
type EarnContract struct {
	ID                string `json:"id"`
	MerchantID        string `json:"merchant_id"`
	PrincipalAmount   string `json:"principal_amount"`
	Currency          string `json:"currency"`
	Blockchain        string `json:"blockchain"`
	Network           string `json:"network"`
	DurationMonths    int    `json:"duration_months"`
	TermDays          int    `json:"term_days"`
	APY               string `json:"apy"`
	DailyInterestAmt  string `json:"daily_interest_amount"`
	AccruedInterest   string `json:"accrued_interest"`
	WithdrawnInterest string `json:"withdrawn_interest"`
	Status            string `json:"status"`
	StartDate         string `json:"start_date"`
	MaturityDate      string `json:"maturity_date"`
	BrokenAt          string `json:"broken_at,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ListEarnContractsParams holds query parameters for listing Earn contracts.
type ListEarnContractsParams struct {
	Status string `json:"status,omitempty"`
	Page   int    `json:"page,omitempty"`
}

// EarnContractList is the response for listing Earn contracts.
type EarnContractList struct {
	Contracts  []EarnContract `json:"contracts"`
	Page       int            `json:"page"`
	TotalCount int            `json:"total_count"`
}

// EarnWithdrawal represents an Earn interest withdrawal.
type EarnWithdrawal struct {
	ID         string `json:"id"`
	ContractID string `json:"contract_id"`
	Amount     string `json:"amount"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	TxHash     string `json:"tx_hash,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ── Refund types ───────────────────────────────────────────────────

// CreateRefundRequest is the request body for creating a refund.
type CreateRefundRequest struct {
	Amount        string `json:"amount"`
	ToAddress     string `json:"to_address"`
	Reason        string `json:"reason"`
	CustomerEmail string `json:"customer_email,omitempty"`
}

// Refund represents a refund.
type Refund struct {
	ID            string `json:"id"`
	InvoiceID     string `json:"invoice_id"`
	MerchantID    string `json:"merchant_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Blockchain    string `json:"blockchain"`
	ToAddress     string `json:"to_address"`
	Reason        string `json:"reason"`
	CustomerEmail string `json:"customer_email,omitempty"`
	Status        string `json:"status"`
	TxHash        string `json:"tx_hash,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ListRefundsParams holds query parameters for listing refunds.
type ListRefundsParams struct {
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// RefundList is the response for listing refunds.
type RefundList struct {
	Refunds    []Refund `json:"refunds"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalCount int      `json:"total_count"`
}

// ── Batch Payout List types ────────────────────────────────────────

// ListBatchPayoutsParams holds query parameters for listing batch payouts.
type ListBatchPayoutsParams struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// BatchPayoutList is the response for listing batch payouts.
type BatchPayoutList struct {
	BatchPayouts []BatchPayout `json:"batch_payouts"`
	Total        int           `json:"total"`
}

// ── Fee Estimate types ────────────────────────────────────────────

// EstimateFeesRequest is the request body for estimating fees.
type EstimateFeesRequest struct {
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Blockchain string `json:"blockchain"`
}

// FeeEstimate is the response for a fee estimation.
type FeeEstimate struct {
	PlatformFee string `json:"platform_fee"`
	GasFee      string `json:"gas_fee"`
	TotalFee    string `json:"total_fee"`
	Currency    string `json:"currency"`
}

// ── Ticket types ──────────────────────────────────────────────────

// CreateTicketRequest is the request body for creating a support ticket.
type CreateTicketRequest struct {
	Subject  string `json:"subject"`
	Message  string `json:"message"`
	Priority string `json:"priority,omitempty"`
}

// Ticket represents a support ticket.
type Ticket struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListTicketsParams holds query parameters for listing tickets.
type ListTicketsParams struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// TicketList is the response for listing tickets.
type TicketList struct {
	Tickets []Ticket `json:"tickets"`
	Total   int      `json:"total"`
}

// TicketReply represents a reply to a support ticket.
type TicketReply struct {
	ID        string `json:"id"`
	TicketID  string `json:"ticket_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// ── Interest History types ────────────────────────────────────────

// InterestEntry represents a single daily interest accrual entry.
type InterestEntry struct {
	AccrualDate        string `json:"accrual_date"`
	DailyInterest      string `json:"daily_interest"`
	CumulativeInterest string `json:"cumulative_interest"`
}

// InterestHistoryList is the response for listing interest history.
type InterestHistoryList struct {
	Entries []InterestEntry `json:"entries"`
}

// ── Balance types ──────────────────────────────────────────────────

// Balance represents a single token balance on a chain.
type Balance struct {
	Chain   string `json:"chain"`
	Token   string `json:"token"`
	Balance string `json:"balance"`
}

// BalanceList is the response for getting balances.
type BalanceList struct {
	Balances []Balance `json:"balances"`
}

// ── Generic list options ─────────────────────────────────────────

// ListOptions holds common query parameters for paginated list endpoints.
type ListOptions struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Status   string `json:"status,omitempty"`
}

// ── Payment Link types ───────────────────────────────────────────

// CreatePaymentLinkRequest is the request body for creating a payment link.
type CreatePaymentLinkRequest struct {
	Amount         string   `json:"amount"`
	Currency       string   `json:"currency,omitempty"`
	Blockchain     string   `json:"blockchain,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	AcceptedTokens []string `json:"accepted_tokens,omitempty"`
	AcceptedChains []string `json:"accepted_chains,omitempty"`
	Title          string   `json:"title,omitempty"`
	Description    string   `json:"description,omitempty"`
	RedirectURL    string   `json:"redirect_url,omitempty"`
	ExternalID     string   `json:"external_id,omitempty"`
}

// UpdatePaymentLinkRequest is the request body for updating a payment link.
type UpdatePaymentLinkRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
	Status      string `json:"status,omitempty"`
}

// PaymentLink represents a payment link.
type PaymentLink struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Blockchain  string `json:"blockchain"`
	Mode        string `json:"mode"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	RedirectURL string `json:"redirect_url,omitempty"`
	ExternalID  string `json:"external_id,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// PaymentLinkList is the response for listing payment links.
type PaymentLinkList struct {
	PaymentLinks []PaymentLink `json:"payment_links"`
	Page         int           `json:"page"`
	PageSize     int           `json:"page_size"`
	TotalCount   int           `json:"total_count"`
}

// ── Subscription Plan types ──────────────────────────────────────

// CreatePlanRequest is the request body for creating a subscription plan.
type CreatePlanRequest struct {
	Name          string `json:"name"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Blockchain    string `json:"blockchain"`
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count,omitempty"`
	Description   string `json:"description,omitempty"`
	TrialDays     int    `json:"trial_days,omitempty"`
}

// SubscriptionPlan represents a subscription plan.
type SubscriptionPlan struct {
	ID            string `json:"id"`
	MerchantID    string `json:"merchant_id"`
	Name          string `json:"name"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Blockchain    string `json:"blockchain"`
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count"`
	Description   string `json:"description,omitempty"`
	TrialDays     int    `json:"trial_days,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// PlanList is the response for listing subscription plans.
type PlanList struct {
	Plans      []SubscriptionPlan `json:"plans"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalCount int                `json:"total_count"`
}

// ── Subscription types ───────────────────────────────────────────

// CreateSubscriptionRequest is the request body for creating a subscription.
type CreateSubscriptionRequest struct {
	PlanID        string                 `json:"plan_id"`
	CustomerEmail string                 `json:"customer_email,omitempty"`
	CustomerName  string                 `json:"customer_name,omitempty"`
	ExternalID    string                 `json:"external_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Subscription represents a subscription.
type Subscription struct {
	ID                 string `json:"id"`
	PlanID             string `json:"plan_id"`
	MerchantID         string `json:"merchant_id"`
	CustomerEmail      string `json:"customer_email,omitempty"`
	CustomerName       string `json:"customer_name,omitempty"`
	ExternalID         string `json:"external_id,omitempty"`
	Status             string `json:"status"`
	CurrentPeriodStart string `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   string `json:"current_period_end,omitempty"`
	CancelledAt        string `json:"cancelled_at,omitempty"`
	PausedAt           string `json:"paused_at,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// SubscriptionResult is the response for creating a subscription (includes invoice).
type SubscriptionResult struct {
	Subscription Subscription `json:"subscription"`
	Invoice      *Invoice     `json:"invoice,omitempty"`
}

// SubscriptionDetail is the response for retrieving a subscription with full details.
type SubscriptionDetail struct {
	Subscription
	Plan     *SubscriptionPlan `json:"plan,omitempty"`
	Invoices []Invoice         `json:"invoices,omitempty"`
}

// SubscriptionList is the response for listing subscriptions.
type SubscriptionList struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	TotalCount    int            `json:"total_count"`
}

// ── Deposit API types ────────────────────────────────────────────

// CreateDepositUserRequest is the request body for creating a deposit user.
type CreateDepositUserRequest struct {
	ExternalID string                 `json:"external_id"`
	Email      string                 `json:"email,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DepositUser represents a deposit API user.
type DepositUser struct {
	ID         string                 `json:"id"`
	MerchantID string                 `json:"merchant_id"`
	ExternalID string                 `json:"external_id"`
	Email      string                 `json:"email,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Status     string                 `json:"status"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

// DepositUserList is the response for listing deposit users.
type DepositUserList struct {
	Users      []DepositUser `json:"users"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalCount int           `json:"total_count"`
}

// DepositAddress represents a blockchain deposit address for a user.
type DepositAddress struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	ExternalID string `json:"external_id"`
	Blockchain string `json:"blockchain"`
	Address    string `json:"address"`
	CreatedAt  string `json:"created_at"`
}

// DepositAddressList is the response for listing deposit addresses.
type DepositAddressList struct {
	Addresses []DepositAddress `json:"addresses"`
}

// DepositEvent represents a deposit event.
type DepositEvent struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	ExternalID string `json:"external_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Blockchain string `json:"blockchain"`
	TxHash     string `json:"tx_hash,omitempty"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// DepositEventList is the response for listing deposit events.
type DepositEventList struct {
	Deposits   []DepositEvent `json:"deposits"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalCount int            `json:"total_count"`
}

// DepositBalance represents a deposit balance for a user on a specific chain/token.
type DepositBalance struct {
	Chain   string `json:"chain"`
	Token   string `json:"token"`
	Balance string `json:"balance"`
}

// DepositBalanceList is the response for listing deposit balances.
type DepositBalanceList struct {
	Balances []DepositBalance `json:"balances"`
}
