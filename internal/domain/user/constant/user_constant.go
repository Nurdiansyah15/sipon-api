package constant

import (
	"time"

	domainerr "sipon-api/internal/domain/errors"
)

const (
	MaxLoginAttempts = 5
	LockoutDuration  = 15 * time.Minute
)

const (
	CodeInvalidEmailFormat        domainerr.Code = "DOMAIN_INVALID_EMAIL_FORMAT"
	CodeInvalidPhoneNumberFormat  domainerr.Code = "DOMAIN_INVALID_PHONE_NUMBER_FORMAT"
	CodeInvalidUsernameFormat     domainerr.Code = "DOMAIN_INVALID_USERNAME_FORMAT"
	CodeInvalidLoginIdentifier    domainerr.Code = "DOMAIN_INVALID_LOGIN_IDENTIFIER"
	CodePasswordTooShort          domainerr.Code = "DOMAIN_PASSWORD_TOO_SHORT"
	CodePasswordMustHaveUppercase domainerr.Code = "DOMAIN_PASSWORD_MUST_HAVE_UPPERCASE"
	CodePasswordMustHaveDigit     domainerr.Code = "DOMAIN_PASSWORD_MUST_HAVE_DIGIT"
	CodeInvalidHashedPassword     domainerr.Code = "DOMAIN_INVALID_HASHED_PASSWORD"
	CodeInvalidOTPFormat          domainerr.Code = "DOMAIN_INVALID_OTP_FORMAT"
	CodeOTPNotNumeric             domainerr.Code = "DOMAIN_OTP_NOT_NUMERIC"

	CodeUserIDRequired              domainerr.Code = "DOMAIN_USER_ID_REQUIRED"
	CodeUserFullnameRequired        domainerr.Code = "DOMAIN_USER_FULLNAME_REQUIRED"
	CodeUserPhoneNumberInvalid      domainerr.Code = "DOMAIN_USER_PHONE_NUMBER_INVALID"
	CodeUserEmailRequired           domainerr.Code = "DOMAIN_USER_EMAIL_REQUIRED"
	CodeUserNotFound                domainerr.Code = "DOMAIN_USER_NOT_FOUND"
	CodeLoginIdentityNotFound       domainerr.Code = "DOMAIN_LOGIN_IDENTITY_NOT_FOUND"
	CodeUserPersistenceFailed       domainerr.Code = "DOMAIN_USER_PERSISTENCE_FAILED"
	CodeUserQueryFailed             domainerr.Code = "DOMAIN_USER_QUERY_FAILED"
	CodeUserDeleted                 domainerr.Code = "DOMAIN_USER_DELETED"
	CodeUserBanned                  domainerr.Code = "DOMAIN_USER_BANNED"
	CodeUserNoCredential            domainerr.Code = "DOMAIN_USER_NO_CREDENTIAL"
	CodeLoginIdentityUnverified     domainerr.Code = "DOMAIN_LOGIN_IDENTITY_UNVERIFIED"
	CodeUserUsernameTaken           domainerr.Code = "DOMAIN_USER_USERNAME_TAKEN"
	CodeUserUsernameSameAsCurrent   domainerr.Code = "DOMAIN_USER_USERNAME_SAME_AS_CURRENT"
	CodeUserNoLocalCredential       domainerr.Code = "DOMAIN_USER_NO_LOCAL_CREDENTIAL"
	CodeUserAlreadyHasLocalPassword domainerr.Code = "DOMAIN_USER_ALREADY_HAS_LOCAL_PASSWORD"
	CodeUserLockedOut               domainerr.Code = "DOMAIN_USER_LOCKED_OUT"
)

type CredentialType string

const (
	CredentialTypeLocal CredentialType = "LOCAL"
)

type LoginIdentifierKind string

const (
	LoginIdentifierEmail    LoginIdentifierKind = "EMAIL"
	LoginIdentifierPhone    LoginIdentifierKind = "PHONE"
	LoginIdentifierUsername LoginIdentifierKind = "USERNAME"
)

type LoginIdentityStatus string

const (
	LoginIdentityStatusVerified   LoginIdentityStatus = "VERIFIED"
	LoginIdentityStatusUnverified LoginIdentityStatus = "UNVERIFIED"
)

type UserStatus string

const (
	StatusActive UserStatus = "ACTIVE"
	StatusBanned UserStatus = "BANNED"
)
