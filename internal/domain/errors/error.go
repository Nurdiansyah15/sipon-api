package errors

import "fmt"

type Code string

type DomainError struct {
	Code Code
	Err  error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %s, err: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("code: %s", e.Code)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func New(code Code) *DomainError {
	return &DomainError{Code: code}
}

func Wrap(code Code, err error) *DomainError {
	return &DomainError{Code: code, Err: err}
}
