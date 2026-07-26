package port

// OTPGenerator abstraksi pembuatan kode OTP acak.
// Implementasi bisa berupa crypto/rand (simple, tidak perlu interface khusus
// tapi dipisah agar mudah di-mock saat testing)
type OTPGenerator interface {
	Generate() (string, error) // mengembalikan 6 digit string
}
