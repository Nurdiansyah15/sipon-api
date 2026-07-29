package valueobject

import (
	"regexp"
	"strings"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
)

var nisRe = regexp.MustCompile(`^1000[12][0-9]{5}$`)

type NIS struct {
	value string
}

func NewNIS(raw string) (NIS, error) {
	v := strings.TrimSpace(raw)
	if !nisRe.MatchString(v) {
		return NIS{}, domainerr.New(constant.CodeInvalidNISFormat)
	}
	return NIS{value: v}, nil
}

func (n NIS) Value() string { return n.value }

func (n NIS) Gender() string {
	if len(n.value) >= 5 {
		return string(n.value[4])
	}
	return ""
}
