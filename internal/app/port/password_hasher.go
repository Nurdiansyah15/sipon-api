package port

// PasswordHasher abstraksi hashing password.
// Implementasi: infrastructure/external/bcrypt
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hashed, plain string) error
}
