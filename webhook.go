package fintrapay

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// VerifyWebhookSignature verifies an FintraPay webhook signature.
//
// rawBody is the raw request body bytes (do NOT parse JSON first).
// signature is the X-FintraPay-Signature header value.
// webhookSecret is your webhook secret from the dashboard.
//
// Returns true if the signature is valid, false otherwise.
//
// Usage with net/http:
//
//	func webhookHandler(w http.ResponseWriter, r *http.Request) {
//	    body, err := io.ReadAll(r.Body)
//	    if err != nil {
//	        http.Error(w, "Bad request", http.StatusBadRequest)
//	        return
//	    }
//	    defer r.Body.Close()
//
//	    sig := r.Header.Get("X-FintraPay-Signature")
//	    if !fintrapay.VerifyWebhookSignature(body, sig, webhookSecret) {
//	        http.Error(w, "Invalid signature", http.StatusUnauthorized)
//	        return
//	    }
//
//	    // Signature valid — parse and process the webhook event
//	    var event map[string]interface{}
//	    json.Unmarshal(body, &event)
//	    // ...
//	}
func VerifyWebhookSignature(rawBody []byte, signature, webhookSecret string) bool {
	if len(rawBody) == 0 || signature == "" || webhookSecret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
