package fonnte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
)

type Sender struct {
	token  string
	url    string
	client *http.Client
}

type sendResponse struct {
	Status bool   `json:"status"`
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

func NewSender(token, url string) *Sender {
	return &Sender{
		token:  strings.TrimSpace(token),
		url:    strings.TrimSpace(url),
		client: http.DefaultClient,
	}
}

func (s *Sender) SendOTP(toPhone, otp string) error {
	msg := fmt.Sprintf("Kode OTP verifikasi nomor HP Anda adalah %s. Berlaku 5 menit.", otp)
	return s.sendRaw(toPhone, msg)
}

func (s *Sender) sendRaw(toPhone, message string) error {
	if s.token == "" {
		return fmt.Errorf("fonnte token belum dikonfigurasi")
	}
	if s.url == "" {
		return fmt.Errorf("fonnte url belum dikonfigurasi")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("target", strings.TrimPrefix(toPhone, "+")); err != nil {
		return fmt.Errorf("fonnte write target: %w", err)
	}
	if err := writer.WriteField("message", message); err != nil {
		return fmt.Errorf("fonnte write message: %w", err)
	}
	if err := writer.WriteField("countryCode", "0"); err != nil {
		return fmt.Errorf("fonnte write country code: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("fonnte finalize body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.url, &body)
	if err != nil {
		return fmt.Errorf("fonnte new request: %w", err)
	}
	req.Header.Set("Authorization", s.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fonnte send request: %w", err)
	}
	defer resp.Body.Close()

	var payload sendResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("fonnte decode response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest || !payload.Status {
		reason := payload.Reason
		if reason == "" {
			reason = payload.Detail
		}
		if reason == "" {
			reason = resp.Status
		}
		return fmt.Errorf("fonnte send failed: %s", reason)
	}
	return nil
}
