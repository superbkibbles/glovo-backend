package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024
	Argon2Threads = 4
	Argon2KeyLen  = 32
	SaltLength    = 32
)

// PasswordHasher handles password hashing and verification using Argon2
type PasswordHasher struct {
	pepper string
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher(pepper string) *PasswordHasher {
	return &PasswordHasher{pepper: pepper}
}

// GenerateSalt generates a random salt
func (p *PasswordHasher) GenerateSalt() (string, error) {
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// HashPassword hashes a password with the given salt
func (p *PasswordHasher) HashPassword(password, salt string) (string, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	pepperedPassword := password + p.pepper
	hash := argon2.IDKey(
		[]byte(pepperedPassword),
		saltBytes,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)

	return base64.StdEncoding.EncodeToString(hash), nil
}

// VerifyPassword verifies a password against a hash
func (p *PasswordHasher) VerifyPassword(password, salt, hash string) (bool, error) {
	computedHash, err := p.HashPassword(password, salt)
	if err != nil {
		return false, err
	}

	computedHashBytes, err := base64.StdEncoding.DecodeString(computedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode computed hash: %w", err)
	}

	hashBytes, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored hash: %w", err)
	}

	return subtle.ConstantTimeCompare(computedHashBytes, hashBytes) == 1, nil
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' ||
			char == '^' || char == '&' || char == '*' || char == '(' || char == ')' ||
			char == '-' || char == '_' || char == '+' || char == '=' || char == '[' ||
			char == ']' || char == '{' || char == '}' || char == '|' || char == '\\' ||
			char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' ||
			char == '>' || char == ',' || char == '.' || char == '?' || char == '/':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
