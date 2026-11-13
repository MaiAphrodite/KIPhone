package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

func errString(s string) error { return errors.New(s) }

func hmacEqual(msg, sig, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(msg)
	return hmac.Equal(mac.Sum(nil), sig)
}
