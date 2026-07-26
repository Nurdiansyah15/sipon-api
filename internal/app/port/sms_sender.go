package port

// SMSSender abstraksi pengiriman pesan ke nomor telepon.
// Implementasi: infrastructure/external/fonnte
type SMSSender interface {
	SendOTP(toPhone, otp string) error
}
