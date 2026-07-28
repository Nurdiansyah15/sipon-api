package valueobject

import (
	"strings"

	domainerr "sipon-api/internal/domain/errors"
)

type UserScopeType string

const (
	ScopeTypeGender UserScopeType = "gender"
)

// Gender scope values
const (
	GenderMale   = "male"
	GenderFemale = "female"
)

func NewUserScopeValue(scopeType UserScopeType, rawValue string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(rawValue))
	switch scopeType {
	case ScopeTypeGender:
		if v != GenderMale && v != GenderFemale {
			return "", domainerr.New("DOMAIN_INVALID_GENDER_SCOPE_VALUE")
		}
		return v, nil
	default:
		return "", domainerr.New("DOMAIN_UNKNOWN_SCOPE_TYPE")
	}
}
