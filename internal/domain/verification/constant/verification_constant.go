package constant

import domainerr "sipon-api/internal/domain/errors"

type CodePurpose string

const (
	PurposeEmailVerification CodePurpose = "EMAIL_VERIFICATION"
	PurposePhoneVerification CodePurpose = "PHONE_VERIFICATION"
	PurposeResetPassword     CodePurpose = "RESET_PASSWORD"
	PurposeChangeEmail       CodePurpose = "CHANGE_EMAIL"
	PurposeChangePhone       CodePurpose = "CHANGE_PHONE"
)

const (
	CodeOTPExpired                    domainerr.Code = "DOMAIN_OTP_EXPIRED"
	CodeOTPUsed                       domainerr.Code = "DOMAIN_OTP_USED"
	CodeOTPInvalid                    domainerr.Code = "DOMAIN_OTP_INVALID"
	CodeVerificationNotFound          domainerr.Code = "DOMAIN_VERIFICATION_NOT_FOUND"
	CodeVerificationQueryFailed       domainerr.Code = "DOMAIN_VERIFICATION_QUERY_FAILED"
	CodeVerificationPersistenceFailed domainerr.Code = "DOMAIN_VERIFICATION_PERSISTENCE_FAILED"
)
