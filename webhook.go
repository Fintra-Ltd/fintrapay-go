package fintrapay

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"
)

// DefaultWebhookMaxAge is the default freshness window for v2 webhook deliveries.
// Anything older (or further in the future) than this is rejected to defeat replay.
const DefaultWebhookMaxAge = 5 * time.Minute

// VerifyWebhookSignature verifies an FintraPay v2 webhook signature.
//
// The v2 envelope signs (timestamp + "\n" + rawBody) with HMAC-SHA256 hex.
// timestamp is the X-FintraPay-Timestamp header (RFC3339). Pass empty string
// only when verifying a legacy v1 delivery (raw-body signing) — discouraged.
//
// Returns true if the signature is valid AND, when timestamp is non-empty,
// the timestamp is within DefaultWebhookMaxAge of now.
//
// Usage with net/http:
//
//	func webhookHandler(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//	    defer r.Body.Close()
//
//	    sig := r.Header.Get("X-FintraPay-Signature")
//	    ts  := r.Header.Get("X-FintraPay-Timestamp")
//	    if !fintrapay.VerifyWebhookSignature(body, sig, webhookSecret, ts) {
//	        http.Error(w, "Invalid signature", http.StatusUnauthorized)
//	        return
//	    }
//	    var event map[string]interface{}
//	    json.Unmarshal(body, &event)
//	    // ...
//	}
func VerifyWebhookSignature(rawBody []byte, signature, webhookSecret, timestamp string) bool {
	return VerifyWebhookSignatureWithMaxAge(rawBody, signature, webhookSecret, timestamp, DefaultWebhookMaxAge)
}

// VerifyWebhookSignatureWithMaxAge is the same as VerifyWebhookSignature but
// lets you tune the freshness window. Use 0 to disable the freshness check.
func VerifyWebhookSignatureWithMaxAge(rawBody []byte, signature, webhookSecret, timestamp string, maxAge time.Duration) bool {
	if len(rawBody) == 0 || signature == "" || webhookSecret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	if timestamp != "" {
		// Freshness check.
		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return false
		}
		if maxAge > 0 {
			delta := time.Since(ts)
			if delta < 0 {
				delta = -delta
			}
			if delta > maxAge {
				return false
			}
		}
		mac.Write([]byte(timestamp + "\n"))
	}
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
