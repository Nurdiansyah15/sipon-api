package port

// EmailSender abstraksi pengiriman email.
// Implementasi: infrastructure/external/smtp
type EmailSender interface {
	SendOTP(toEmail, username, otp string) error
	// SendPasswordResetOTP mengirim OTP khusus konteks reset password — beda body/copy
	// dari SendOTP (yang khusus verifikasi email saat registrasi), supaya user tidak
	// menerima email yang menyebut "registrasi" saat sedang reset password.
	SendPasswordResetOTP(toEmail, username, otp string) error
}
