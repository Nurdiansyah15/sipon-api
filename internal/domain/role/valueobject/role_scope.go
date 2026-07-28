package valueobject

import (
	"strings"

	domainerr "sipon-api/internal/domain/errors"
)

type ScopeType string

const (
	ScopeTypeGender ScopeType = "gender"
)

const (
	GenderMale   = "male"
	GenderFemale = "female"
)

func NewScopeValue(scopeType ScopeType, rawValue string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(rawValue))
	switch scopeType {
	case ScopeTypeGender:
		if v != GenderMale && v != GenderFemale {
			return "", domainerr.New("DOMAIN_INVALID_SCOPE_VALUE")
		}
		return v, nil
	default:
		return "", domainerr.New("DOMAIN_UNKNOWN_SCOPE_TYPE")
	}
}
