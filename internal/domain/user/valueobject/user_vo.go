package valueobject

import (
	"regexp"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/constant"
	"strings"
	"unicode"
)

// ── Email ────────────────────────────────────────────────────────────────────

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var phoneNumberRe = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
var nisPatternRe = regexp.MustCompile(`^1000[12][0-9]{5}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if !emailRe.MatchString(v) {
		return Email{}, domainerr.New(constant.CodeInvalidEmailFormat)
	}
	return Email{value: v}, nil
}

func (e Email) Value() string  { return e.value }
func (e Email) String() string { return e.value }

// ── PhoneNumber ─────────────────────────────────────────────────────────────

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(raw string) (*PhoneNumber, error) {
	var digits []rune
	for _, r := range strings.TrimSpace(raw) {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) < 8 || len(digits) > 15 {
		return nil, domainerr.New(constant.CodeInvalidPhoneNumberFormat)
	}

	v := string(digits)

	switch {
	case strings.HasPrefix(v, "0"):
		v = "62" + v[1:]

	case strings.HasPrefix(v, "62"):
		// sudah benar, tidak perlu diubah
	default:
		v = "62" + v
	}

	v = "+" + v

	if !phoneNumberRe.MatchString(v) {
		return nil, domainerr.New(constant.CodeInvalidPhoneNumberFormat)
	}
	return &PhoneNumber{value: v}, nil
}

func (p *PhoneNumber) Value() string  { return p.value }
func (p *PhoneNumber) String() string { return p.value }

// ── HashedPassword ───────────────────────────────────────────────────────────

type HashedPassword struct {
	value string
}

func NewHashedPassword(hash string) (HashedPassword, error) {
	if len(hash) < 10 {
		return HashedPassword{}, domainerr.New(constant.CodeInvalidHashedPassword)
	}
	return HashedPassword{value: hash}, nil
}

func (h HashedPassword) Value() string { return h.value }

// ── PlainPassword (hanya untuk validasi kekuatan, tidak disimpan) ────────────

type PlainPassword struct {
	value string
}

func NewPlainPassword(raw string) (PlainPassword, error) {
	if len(raw) < 8 {
		return PlainPassword{}, domainerr.New(constant.CodePasswordTooShort)
	}
	var hasUpper, hasDigit bool
	for _, r := range raw {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasUpper {
		return PlainPassword{}, domainerr.New(constant.CodePasswordMustHaveUppercase)
	}
	if !hasDigit {
		return PlainPassword{}, domainerr.New(constant.CodePasswordMustHaveDigit)
	}
	return PlainPassword{value: raw}, nil
}

func (p PlainPassword) Value() string { return p.value }

// ── OTPCode ──────────────────────────────────────────────────────────────────

type OTPCode struct {
	value string
}

func NewOTPCode(code string) (OTPCode, error) {
	if len(code) != 6 {
		return OTPCode{}, domainerr.New(constant.CodeInvalidOTPFormat)
	}
	for _, r := range code {
		if !unicode.IsDigit(r) {
			return OTPCode{}, domainerr.New(constant.CodeOTPNotNumeric)
		}
	}
	return OTPCode{value: code}, nil
}

func (o OTPCode) Value() string           { return o.value }
func (o OTPCode) Match(input string) bool { return o.value == input }

// ── Username ─────────────────────────────────────────────────────────────────

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type Username struct {
	value string
}

func NewUsername(raw string) (Username, error) {
	v := strings.TrimSpace(raw)
	if !usernameRe.MatchString(v) {
		return Username{}, domainerr.New(constant.CodeInvalidUsernameFormat)
	}
	return Username{value: v}, nil
}

func (u Username) Value() string { return u.value }

// ── LoginIdentifier ─────────────────────────────────────────────────────────

type LoginIdentifier struct {
	kind  constant.LoginIdentifierKind
	value string
}

func NewLoginIdentifier(raw string) (LoginIdentifier, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return LoginIdentifier{}, domainerr.New(constant.CodeInvalidLoginIdentifier)
	}

	if email, err := NewEmail(v); err == nil {
		return LoginIdentifier{kind: constant.LoginIdentifierEmail, value: email.Value()}, nil
	}
	if phone, err := NewPhoneNumber(v); err == nil {
		return LoginIdentifier{kind: constant.LoginIdentifierPhone, value: phone.Value()}, nil
	}
	if nisPatternRe.MatchString(v) {
		return LoginIdentifier{kind: constant.LoginIdentifierNIS, value: v}, nil
	}
	if username, err := NewUsername(v); err == nil {
		return LoginIdentifier{kind: constant.LoginIdentifierUsername, value: username.Value()}, nil
	}

	return LoginIdentifier{}, domainerr.New(constant.CodeInvalidLoginIdentifier)
}

func (i LoginIdentifier) Kind() constant.LoginIdentifierKind { return i.kind }
func (i LoginIdentifier) Value() string                      { return i.value }
