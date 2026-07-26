package otpgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// CryptoOTPGenerator mengimplementasi port.OTPGenerator
type CryptoOTPGenerator struct{}

func NewCryptoOTPGenerator() *CryptoOTPGenerator {
	return &CryptoOTPGenerator{}
}

func (g *CryptoOTPGenerator) Generate() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	// Pastikan selalu 6 digit dengan leading zero
	return fmt.Sprintf("%06d", n.Int64()), nil
}
