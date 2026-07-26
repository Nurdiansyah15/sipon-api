package smtp

import (
	"log/slog"
)

type NoopEmailSender struct{}

func NewNoopEmailSender() *NoopEmailSender {
	return &NoopEmailSender{}
}

func (n *NoopEmailSender) SendOTP(toEmail, username, otp string) error {
	slog.Info("noop email: SendOTP", slog.String("to", toEmail), slog.String("username", username))
	return nil
}

func (n *NoopEmailSender) SendPasswordResetOTP(toEmail, username, otp string) error {
	slog.Info("noop email: SendPasswordResetOTP", slog.String("to", toEmail), slog.String("username", username))
	return nil
}
