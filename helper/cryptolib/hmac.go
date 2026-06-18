package cryptolib

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HMACInterface interface {
	Hash(value string) string
}

type hmacHelper struct{ key []byte }

func NewHMAC(key []byte) HMACInterface { return &hmacHelper{key: key} }

func (h *hmacHelper) Hash(value string) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
