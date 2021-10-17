// Package encode signs HMAC SHA25 for binance request
package encode

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Signer signs provided payloads.
type Signer interface {
	// Sign signs provided payload and returns encoded string sum.
	Sign(payload []byte) (string, error)
}

// Hmac uses HMAC SHA256 for signing payloads.
type Hmac struct {
	Key []byte
}

// Sign signs provided payload and returns encoded string sum.
func (hs *Hmac) Sign(payload []byte) (s string, err error) {
	mac := hmac.New(sha256.New, hs.Key)
	_, err = mac.Write(payload)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}
