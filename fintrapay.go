// Package fintrapay provides a Go client for the FintraPay API
// with automatic HMAC-SHA256 request signing.
//
// Usage:
//
//	client := fintrapay.NewClient("xfp_key_...", "xfp_secret_...")
//	invoice, err := client.CreateInvoice(ctx, &fintrapay.CreateInvoiceRequest{
//	    Amount:     "100.00",
//	    Currency:   "USDT",
//	    Blockchain: "tron",
//	})
package fintrapay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://fintrapay.io/v1"
const defaultTimeout = 30 * time.Second

// ── Error types ────────────────────────────────────────────────────

// APIError is the base error type for FintraPay API errors.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("fintrapay: %s (code=%s, status=%d)", e.Message, e.Code, e.StatusCode)
	}
	return fmt.Sprintf("fintrapay: %s (status=%d)", e.Message, e.StatusCode)
}

// AuthError is returned when API authentication fails (HTTP 401).
type AuthError struct {
	APIError
}

// ValidationError is returned when request validation fails (HTTP 422).
type ValidationError struct {
	APIError
}

// RateLimitError is returned when the rate limit is exceeded (HTTP 429).
type RateLimitError struct {
	APIError
	RetryAfter int // seconds to wait before retrying
}

// ── Client ─────────────────────────────────────────────────────────

// Client is the FintraPay API client.
type Client struct {
	APIKey     string
	APISecret  string
	BaseURL    string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewClient creates a new FintraPay API client.
//
// apiKey and apiSecret are required. Use Option functions to customise
// the base URL, timeout, or HTTP client.
func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   defaultBaseURL,
		Timeout:   defaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: c.Timeout}
	}
	return c
}

// ── Request signing ────────────────────────────────────────────────

// sign computes the HMAC-SHA256 signature and returns the auth headers.
func (c *Client) sign(method, path, body string) map[string]string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := timestamp + "\n" + method + "\n" + path + "\n" + body

	mac := hmac.New(sha256.New, []byte(c.APISecret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return map[string]string{
		"X-API-Key":    c.APIKey,
		"X-Timestamp":  timestamp,
		"X-Signature":  signature,
		"Content-Type": "application/json",
	}
}

// ── Internal HTTP ──────────────────────────────────────────────────

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (json.RawMessage, error) {
	var bodyStr string
	var bodyReader io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("fintrapay: failed to marshal request body: %w", err)
		}
		bodyStr = string(b)
		bodyReader = bytes.NewReader(b)
	}

	fullURL := c.BaseURL + path
	headers := c.sign(method, path, bodyStr)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("fintrapay: failed to create request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fintrapay: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fintrapay: failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleError(resp, respBody)
	}

	return json.RawMessage(respBody), nil
}

func (c *Client) handleError(resp *http.Response, body []byte) error {
	var errResp struct {
		ErrorMsg string                 `json:"error"`
		Code     string                 `json:"code"`
		Details  map[string]interface{} `json:"details"`
	}
	_ = json.Unmarshal(body, &errResp)

	if errResp.ErrorMsg == "" {
		errResp.ErrorMsg = "Unknown error"
	}

	base := APIError{
		StatusCode: resp.StatusCode,
		Code:       errResp.Code,
		Message:    errResp.ErrorMsg,
		Details:    errResp.Details,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthError{APIError: base}
	case http.StatusUnprocessableEntity:
		return &ValidationError{APIError: base}
	case http.StatusTooManyRequests:
		retryAfter := 60
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if v, err := strconv.Atoi(ra); err == nil {
				retryAfter = v
			}
		}
		return &RateLimitError{APIError: base, RetryAfter: retryAfter}
	default:
		return &base
	}
}

// decode unmarshals a JSON response into the target.
func decode(raw json.RawMessage, target interface{}) error {
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("fintrapay: failed to decode response: %w", err)
	}
	return nil
}

// ── Invoices ───────────────────────────────────────────────────────

// CreateInvoice creates a payment invoice.
//
// For single-token invoices, set Currency and Blockchain.
// For multi-token invoices, set AcceptedTokens and AcceptedChains instead.
func (c *Client) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*Invoice, error) {
	raw, err := c.doRequest(ctx, "POST", "/invoices", req)
	if err != nil {
		return nil, err
	}
	var inv Invoice
	return &inv, decode(raw, &inv)
}

// GetInvoice retrieves an invoice by ID.
func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	raw, err := c.doRequest(ctx, "GET", "/invoices/"+invoiceID, nil)
	if err != nil {
		return nil, err
	}
	var inv Invoice
	return &inv, decode(raw, &inv)
}

// ListInvoices lists invoices with optional filters.
func (c *Client) ListInvoices(ctx context.Context, params *ListInvoicesParams) (*InvoiceList, error) {
	path := "/invoices" + buildInvoiceQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list InvoiceList
	return &list, decode(raw, &list)
}

func buildInvoiceQuery(p *ListInvoicesParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Blockchain != "" {
		q.Set("blockchain", p.Blockchain)
	}
	if p.Currency != "" {
		q.Set("currency", p.Currency)
	}
	if p.Mode != "" {
		q.Set("mode", p.Mode)
	}
	return "?" + q.Encode()
}

// ── Payouts ────────────────────────────────────────────────────────

// CreatePayout creates a single payout to any address.
func (c *Client) CreatePayout(ctx context.Context, req *CreatePayoutRequest) (*Payout, error) {
	raw, err := c.doRequest(ctx, "POST", "/payouts", req)
	if err != nil {
		return nil, err
	}
	var p Payout
	return &p, decode(raw, &p)
}

// CreateBatchPayout creates a batch payout with multiple recipients.
func (c *Client) CreateBatchPayout(ctx context.Context, req *CreateBatchPayoutRequest) (*BatchPayout, error) {
	raw, err := c.doRequest(ctx, "POST", "/payouts/batch", req)
	if err != nil {
		return nil, err
	}
	var bp BatchPayout
	return &bp, decode(raw, &bp)
}

// GetPayout retrieves a payout by ID.
func (c *Client) GetPayout(ctx context.Context, payoutID string) (*Payout, error) {
	raw, err := c.doRequest(ctx, "GET", "/payouts/"+payoutID, nil)
	if err != nil {
		return nil, err
	}
	var p Payout
	return &p, decode(raw, &p)
}

// ListPayouts lists payouts with optional filters.
func (c *Client) ListPayouts(ctx context.Context, params *ListPayoutsParams) (*PayoutList, error) {
	path := "/payouts" + buildPayoutQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list PayoutList
	return &list, decode(raw, &list)
}

func buildPayoutQuery(p *ListPayoutsParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	return "?" + q.Encode()
}

// ListBatchPayouts lists batch payouts with optional pagination.
func (c *Client) ListBatchPayouts(ctx context.Context, params *ListBatchPayoutsParams) (*BatchPayoutList, error) {
	path := "/payouts/batches" + buildBatchPayoutQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list BatchPayoutList
	return &list, decode(raw, &list)
}

// GetBatchPayout retrieves a batch payout by ID.
func (c *Client) GetBatchPayout(ctx context.Context, batchID string) (*BatchPayout, error) {
	raw, err := c.doRequest(ctx, "GET", "/payouts/batches/"+batchID, nil)
	if err != nil {
		return nil, err
	}
	var bp BatchPayout
	return &bp, decode(raw, &bp)
}

func buildBatchPayoutQuery(p *ListBatchPayoutsParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	return "?" + q.Encode()
}

// ── Withdrawals ────────────────────────────────────────────────────

// CreateWithdrawal creates a withdrawal to the merchant's registered wallet.
func (c *Client) CreateWithdrawal(ctx context.Context, req *CreateWithdrawalRequest) (*Withdrawal, error) {
	raw, err := c.doRequest(ctx, "POST", "/withdrawals", req)
	if err != nil {
		return nil, err
	}
	var w Withdrawal
	return &w, decode(raw, &w)
}

// GetWithdrawal retrieves a withdrawal by ID.
func (c *Client) GetWithdrawal(ctx context.Context, withdrawalID string) (*Withdrawal, error) {
	raw, err := c.doRequest(ctx, "GET", "/withdrawals/"+withdrawalID, nil)
	if err != nil {
		return nil, err
	}
	var w Withdrawal
	return &w, decode(raw, &w)
}

// ListWithdrawals lists withdrawals with optional pagination.
func (c *Client) ListWithdrawals(ctx context.Context, params *ListWithdrawalsParams) (*WithdrawalList, error) {
	path := "/withdrawals" + buildWithdrawalQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list WithdrawalList
	return &list, decode(raw, &list)
}

func buildWithdrawalQuery(p *ListWithdrawalsParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	return "?" + q.Encode()
}

// ── Earn ───────────────────────────────────────────────────────────

// CreateEarnContract creates a new Earn contract.
func (c *Client) CreateEarnContract(ctx context.Context, req *CreateEarnContractRequest) (*EarnContract, error) {
	raw, err := c.doRequest(ctx, "POST", "/earn/contracts", req)
	if err != nil {
		return nil, err
	}
	var ec EarnContract
	return &ec, decode(raw, &ec)
}

// GetEarnContract retrieves an Earn contract by ID.
func (c *Client) GetEarnContract(ctx context.Context, contractID string) (*EarnContract, error) {
	raw, err := c.doRequest(ctx, "GET", "/earn/contracts/"+contractID, nil)
	if err != nil {
		return nil, err
	}
	var ec EarnContract
	return &ec, decode(raw, &ec)
}

// ListEarnContracts lists Earn contracts with optional filters.
func (c *Client) ListEarnContracts(ctx context.Context, params *ListEarnContractsParams) (*EarnContractList, error) {
	path := "/earn/contracts" + buildEarnQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list EarnContractList
	return &list, decode(raw, &list)
}

func buildEarnQuery(p *ListEarnContractsParams) string {
	if p == nil {
		return "?page=1"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	q.Set("page", strconv.Itoa(page))
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	return "?" + q.Encode()
}

// WithdrawEarnInterest withdraws accrued interest from an Earn contract (minimum $10).
func (c *Client) WithdrawEarnInterest(ctx context.Context, contractID, amount string) (*EarnWithdrawal, error) {
	body := map[string]string{"amount": amount}
	raw, err := c.doRequest(ctx, "POST", "/earn/contracts/"+contractID+"/withdraw-interest", body)
	if err != nil {
		return nil, err
	}
	var ew EarnWithdrawal
	return &ew, decode(raw, &ew)
}

// BreakEarnContract performs an early break of an Earn contract.
func (c *Client) BreakEarnContract(ctx context.Context, contractID string) (*EarnContract, error) {
	raw, err := c.doRequest(ctx, "POST", "/earn/contracts/"+contractID+"/break", nil)
	if err != nil {
		return nil, err
	}
	var ec EarnContract
	return &ec, decode(raw, &ec)
}

// ── Refunds ────────────────────────────────────────────────────────

// CreateRefund creates a refund for a paid invoice. Partial refunds are supported.
func (c *Client) CreateRefund(ctx context.Context, invoiceID string, req *CreateRefundRequest) (*Refund, error) {
	raw, err := c.doRequest(ctx, "POST", "/invoices/"+invoiceID+"/refunds", req)
	if err != nil {
		return nil, err
	}
	var r Refund
	return &r, decode(raw, &r)
}

// GetRefund retrieves a refund by ID.
func (c *Client) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	raw, err := c.doRequest(ctx, "GET", "/refunds/"+refundID, nil)
	if err != nil {
		return nil, err
	}
	var r Refund
	return &r, decode(raw, &r)
}

// ListRefunds lists all refunds with optional filters.
func (c *Client) ListRefunds(ctx context.Context, params *ListRefundsParams) (*RefundList, error) {
	path := "/refunds" + buildRefundQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list RefundList
	return &list, decode(raw, &list)
}

func buildRefundQuery(p *ListRefundsParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	return "?" + q.Encode()
}

// ListInvoiceRefunds lists all refunds for a specific invoice.
func (c *Client) ListInvoiceRefunds(ctx context.Context, invoiceID string) (*RefundList, error) {
	raw, err := c.doRequest(ctx, "GET", "/invoices/"+invoiceID+"/refunds", nil)
	if err != nil {
		return nil, err
	}
	var list RefundList
	return &list, decode(raw, &list)
}

// ── Fees ──────────────────────────────────────────────────────────

// EstimateFees estimates platform and gas fees for a transaction.
func (c *Client) EstimateFees(ctx context.Context, req *EstimateFeesRequest) (*FeeEstimate, error) {
	raw, err := c.doRequest(ctx, "POST", "/fees/estimate", req)
	if err != nil {
		return nil, err
	}
	var fe FeeEstimate
	return &fe, decode(raw, &fe)
}

// ── Tickets ───────────────────────────────────────────────────────

// CreateTicket creates a new support ticket.
func (c *Client) CreateTicket(ctx context.Context, req *CreateTicketRequest) (*Ticket, error) {
	raw, err := c.doRequest(ctx, "POST", "/tickets", req)
	if err != nil {
		return nil, err
	}
	var t Ticket
	return &t, decode(raw, &t)
}

// ListTickets lists support tickets with optional pagination.
func (c *Client) ListTickets(ctx context.Context, params *ListTicketsParams) (*TicketList, error) {
	path := "/tickets" + buildTicketQuery(params)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list TicketList
	return &list, decode(raw, &list)
}

// GetTicket retrieves a support ticket by ID.
func (c *Client) GetTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	raw, err := c.doRequest(ctx, "GET", "/tickets/"+ticketID, nil)
	if err != nil {
		return nil, err
	}
	var t Ticket
	return &t, decode(raw, &t)
}

// ReplyTicket adds a reply to an existing support ticket.
func (c *Client) ReplyTicket(ctx context.Context, ticketID string, message string) (*TicketReply, error) {
	body := map[string]string{"message": message}
	raw, err := c.doRequest(ctx, "POST", "/tickets/"+ticketID+"/reply", body)
	if err != nil {
		return nil, err
	}
	var tr TicketReply
	return &tr, decode(raw, &tr)
}

func buildTicketQuery(p *ListTicketsParams) string {
	if p == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	return "?" + q.Encode()
}

// ── Interest History ──────────────────────────────────────────────

// GetInterestHistory retrieves the interest accrual history for an Earn contract.
func (c *Client) GetInterestHistory(ctx context.Context, contractID string) (*InterestHistoryList, error) {
	raw, err := c.doRequest(ctx, "GET", "/earn/contracts/"+contractID+"/interest-history", nil)
	if err != nil {
		return nil, err
	}
	var list InterestHistoryList
	return &list, decode(raw, &list)
}

// ── Balance ────────────────────────────────────────────────────────

// GetBalance retrieves custodial balances across all chains.
func (c *Client) GetBalance(ctx context.Context) (*BalanceList, error) {
	raw, err := c.doRequest(ctx, "GET", "/balance", nil)
	if err != nil {
		return nil, err
	}
	var bl BalanceList
	return &bl, decode(raw, &bl)
}

// ── Payment Links ─────────────────────────────────────────────────

// CreatePaymentLink creates a new payment link.
func (c *Client) CreatePaymentLink(ctx context.Context, req *CreatePaymentLinkRequest) (*PaymentLink, error) {
	raw, err := c.doRequest(ctx, "POST", "/payment-links", req)
	if err != nil {
		return nil, err
	}
	var pl PaymentLink
	return &pl, decode(raw, &pl)
}

// ListPaymentLinks lists payment links with optional filters.
func (c *Client) ListPaymentLinks(ctx context.Context, opts *ListOptions) (*PaymentLinkList, error) {
	path := "/payment-links" + buildListQuery(opts)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list PaymentLinkList
	return &list, decode(raw, &list)
}

// GetPaymentLink retrieves a payment link by ID.
func (c *Client) GetPaymentLink(ctx context.Context, linkID string) (*PaymentLink, error) {
	raw, err := c.doRequest(ctx, "GET", "/payment-links/"+linkID, nil)
	if err != nil {
		return nil, err
	}
	var pl PaymentLink
	return &pl, decode(raw, &pl)
}

// UpdatePaymentLink updates an existing payment link.
func (c *Client) UpdatePaymentLink(ctx context.Context, linkID string, req *UpdatePaymentLinkRequest) (*PaymentLink, error) {
	raw, err := c.doRequest(ctx, "PATCH", "/payment-links/"+linkID, req)
	if err != nil {
		return nil, err
	}
	var pl PaymentLink
	return &pl, decode(raw, &pl)
}

// ── Subscription Plans ────────────────────────────────────────────

// CreateSubscriptionPlan creates a new subscription plan.
func (c *Client) CreateSubscriptionPlan(ctx context.Context, req *CreatePlanRequest) (*SubscriptionPlan, error) {
	raw, err := c.doRequest(ctx, "POST", "/subscription-plans", req)
	if err != nil {
		return nil, err
	}
	var sp SubscriptionPlan
	return &sp, decode(raw, &sp)
}

// ListSubscriptionPlans lists subscription plans with optional filters.
func (c *Client) ListSubscriptionPlans(ctx context.Context, opts *ListOptions) (*PlanList, error) {
	path := "/subscription-plans" + buildListQuery(opts)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list PlanList
	return &list, decode(raw, &list)
}

// GetSubscriptionPlan retrieves a subscription plan by ID.
func (c *Client) GetSubscriptionPlan(ctx context.Context, planID string) (*SubscriptionPlan, error) {
	raw, err := c.doRequest(ctx, "GET", "/subscription-plans/"+planID, nil)
	if err != nil {
		return nil, err
	}
	var sp SubscriptionPlan
	return &sp, decode(raw, &sp)
}

// UpdateSubscriptionPlan updates an existing subscription plan.
func (c *Client) UpdateSubscriptionPlan(ctx context.Context, planID string, req map[string]interface{}) (*SubscriptionPlan, error) {
	raw, err := c.doRequest(ctx, "PATCH", "/subscription-plans/"+planID, req)
	if err != nil {
		return nil, err
	}
	var sp SubscriptionPlan
	return &sp, decode(raw, &sp)
}

// ── Subscriptions ─────────────────────────────────────────────────

// CreateSubscription creates a new subscription.
func (c *Client) CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*SubscriptionResult, error) {
	raw, err := c.doRequest(ctx, "POST", "/subscriptions", req)
	if err != nil {
		return nil, err
	}
	var sr SubscriptionResult
	return &sr, decode(raw, &sr)
}

// ListSubscriptions lists subscriptions with optional filters.
func (c *Client) ListSubscriptions(ctx context.Context, opts *ListOptions) (*SubscriptionList, error) {
	path := "/subscriptions" + buildListQuery(opts)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list SubscriptionList
	return &list, decode(raw, &list)
}

// GetSubscription retrieves a subscription by ID.
func (c *Client) GetSubscription(ctx context.Context, subID string) (*SubscriptionDetail, error) {
	raw, err := c.doRequest(ctx, "GET", "/subscriptions/"+subID, nil)
	if err != nil {
		return nil, err
	}
	var sd SubscriptionDetail
	return &sd, decode(raw, &sd)
}

// CancelSubscription cancels a subscription.
func (c *Client) CancelSubscription(ctx context.Context, subID string, reason string) (*Subscription, error) {
	body := map[string]string{"reason": reason}
	raw, err := c.doRequest(ctx, "POST", "/subscriptions/"+subID+"/cancel", body)
	if err != nil {
		return nil, err
	}
	var s Subscription
	return &s, decode(raw, &s)
}

// PauseSubscription pauses an active subscription.
func (c *Client) PauseSubscription(ctx context.Context, subID string) (*Subscription, error) {
	raw, err := c.doRequest(ctx, "POST", "/subscriptions/"+subID+"/pause", nil)
	if err != nil {
		return nil, err
	}
	var s Subscription
	return &s, decode(raw, &s)
}

// ResumeSubscription resumes a paused subscription.
func (c *Client) ResumeSubscription(ctx context.Context, subID string) (*Subscription, error) {
	raw, err := c.doRequest(ctx, "POST", "/subscriptions/"+subID+"/resume", nil)
	if err != nil {
		return nil, err
	}
	var s Subscription
	return &s, decode(raw, &s)
}

// ── Deposit API ───────────────────────────────────────────────────

// CreateDepositUser creates a new deposit API user.
func (c *Client) CreateDepositUser(ctx context.Context, req *CreateDepositUserRequest) (*DepositUser, error) {
	raw, err := c.doRequest(ctx, "POST", "/deposit-api/users", req)
	if err != nil {
		return nil, err
	}
	var du DepositUser
	return &du, decode(raw, &du)
}

// GetDepositUser retrieves a deposit user by external ID.
func (c *Client) GetDepositUser(ctx context.Context, extID string) (*DepositUser, error) {
	raw, err := c.doRequest(ctx, "GET", "/deposit-api/users/"+extID, nil)
	if err != nil {
		return nil, err
	}
	var du DepositUser
	return &du, decode(raw, &du)
}

// ListDepositUsers lists deposit users with optional pagination.
func (c *Client) ListDepositUsers(ctx context.Context, opts *ListOptions) (*DepositUserList, error) {
	path := "/deposit-api/users" + buildListQuery(opts)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list DepositUserList
	return &list, decode(raw, &list)
}

// UpdateDepositUser updates an existing deposit user.
func (c *Client) UpdateDepositUser(ctx context.Context, extID string, req map[string]interface{}) (*DepositUser, error) {
	raw, err := c.doRequest(ctx, "PATCH", "/deposit-api/users/"+extID, req)
	if err != nil {
		return nil, err
	}
	var du DepositUser
	return &du, decode(raw, &du)
}

// CreateDepositAddress creates a deposit address for a specific blockchain.
func (c *Client) CreateDepositAddress(ctx context.Context, extID string, blockchain string) (*DepositAddress, error) {
	body := map[string]string{"blockchain": blockchain}
	raw, err := c.doRequest(ctx, "POST", "/deposit-api/users/"+extID+"/addresses", body)
	if err != nil {
		return nil, err
	}
	var da DepositAddress
	return &da, decode(raw, &da)
}

// CreateAllDepositAddresses creates deposit addresses on all supported blockchains.
func (c *Client) CreateAllDepositAddresses(ctx context.Context, extID string) (*DepositAddressList, error) {
	raw, err := c.doRequest(ctx, "POST", "/deposit-api/users/"+extID+"/addresses/all", nil)
	if err != nil {
		return nil, err
	}
	var list DepositAddressList
	return &list, decode(raw, &list)
}

// ListDepositAddresses lists all deposit addresses for a user.
func (c *Client) ListDepositAddresses(ctx context.Context, extID string) (*DepositAddressList, error) {
	raw, err := c.doRequest(ctx, "GET", "/deposit-api/users/"+extID+"/addresses", nil)
	if err != nil {
		return nil, err
	}
	var list DepositAddressList
	return &list, decode(raw, &list)
}

// ListDeposits lists deposit events with optional filters.
func (c *Client) ListDeposits(ctx context.Context, opts *ListOptions) (*DepositEventList, error) {
	path := "/deposit-api/deposits" + buildListQuery(opts)
	raw, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var list DepositEventList
	return &list, decode(raw, &list)
}

// GetDeposit retrieves a deposit event by ID.
func (c *Client) GetDeposit(ctx context.Context, depositID string) (*DepositEvent, error) {
	raw, err := c.doRequest(ctx, "GET", "/deposit-api/deposits/"+depositID, nil)
	if err != nil {
		return nil, err
	}
	var de DepositEvent
	return &de, decode(raw, &de)
}

// ListDepositBalances lists deposit balances for a user.
func (c *Client) ListDepositBalances(ctx context.Context, extID string) (*DepositBalanceList, error) {
	raw, err := c.doRequest(ctx, "GET", "/deposit-api/users/"+extID+"/balances", nil)
	if err != nil {
		return nil, err
	}
	var list DepositBalanceList
	return &list, decode(raw, &list)
}

// ── Generic list query builder ────────────────────────────────────

func buildListQuery(opts *ListOptions) string {
	if opts == nil {
		return "?page=1&page_size=20"
	}
	q := url.Values{}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	return "?" + q.Encode()
}
