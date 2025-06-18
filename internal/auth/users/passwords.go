package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// generate a random salt of n bytes
func GenSalt(n int) []byte {
	salt := make([]byte, n)
	_, err := rand.Read(salt)
	if err != nil {
		return nil
	}
	return salt
}

// hash password using Argon2id with salt
func HashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	// Encode salt and hash for storage
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("%s.%s", b64Salt, b64Hash)
}

// checks for length and character diversity on password input
func ValidatePassword(password string) error {
	// more will likely be added to this validate function
	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
		hasSymbol bool
	)

	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSymbol = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSymbol {
		return fmt.Errorf("password must contain at least one symbol")
	}

	return nil
}

// compare password to stored hash
func VerifyPassword(password string, fullHash string) bool {
	parts := strings.Split(fullHash, ".")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expectedHash := HashPassword(password, salt)
	return expectedHash == fullHash
}
