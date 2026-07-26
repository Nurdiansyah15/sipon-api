package smtp

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
)

type SMTPEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPEmailSender(host, port, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailSender) SendOTP(toEmail, username, otp string) error {
	headerFrom, envelopeFrom := resolveSenderAddress(s.from, s.username)

	subject := "Kode Verifikasi Email Anda"
	body := fmt.Sprintf(`Halo %s,

Kode OTP verifikasi email Anda adalah:

    %s

Kode ini berlaku selama 5 menit.
Jika Anda tidak melakukan registrasi, abaikan email ini.

Salam,
Tim sipon-api`, username, otp)

	msg := buildMsg(headerFrom, toEmail, subject, body)
	return s.sendRaw(envelopeFrom, toEmail, msg)
}

func (s *SMTPEmailSender) SendPasswordResetOTP(toEmail, username, otp string) error {
	headerFrom, envelopeFrom := resolveSenderAddress(s.from, s.username)

	subject := "Kode Reset Password Anda"
	body := fmt.Sprintf(`Halo %s,

Kami menerima permintaan reset password untuk akun Anda. Kode OTP-nya adalah:

    %s

Kode ini berlaku selama 5 menit.
Jika Anda tidak meminta reset password, abaikan email ini — password Anda tidak akan berubah.

Salam,
Tim sipon-api`, username, otp)

	msg := buildMsg(headerFrom, toEmail, subject, body)
	return s.sendRaw(envelopeFrom, toEmail, msg)
}

// sendRaw mengirim raw SMTP message. Mendukung TLS (port 465) dan STARTTLS (587).
func (s *SMTPEmailSender) sendRaw(envelopeFrom, toEmail, msg string) error {
	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	if s.port == "465" {
		tlsConfig := &tls.Config{ServerName: s.host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err = client.Mail(envelopeFrom); err != nil {
			return err
		}
		if err = client.Rcpt(toEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err = fmt.Fprint(w, msg); err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, envelopeFrom, []string{toEmail}, []byte(msg))
}

func buildMsg(from, to, subject, body string) string {
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	)
}

func resolveSenderAddress(configuredFrom, username string) (headerFrom string, envelopeFrom string) {
	fromValue := strings.TrimSpace(configuredFrom)
	userValue := strings.TrimSpace(username)

	fromAddr := parseEmailAddress(fromValue)
	userAddr := parseEmailAddress(userValue)

	envelopeFrom = userAddr
	if envelopeFrom == "" {
		envelopeFrom = userValue
	}

	headerFrom = fromAddr
	if headerFrom == "" {
		headerFrom = envelopeFrom
	}
	if headerFrom == "" {
		headerFrom = "noreply@example.com"
	}

	return headerFrom, envelopeFrom
}

func parseEmailAddress(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if addr, err := mail.ParseAddress(value); err == nil {
		return strings.TrimSpace(addr.Address)
	}
	if strings.Count(value, "@") == 1 {
		return strings.TrimSpace(value)
	}
	return ""
}
