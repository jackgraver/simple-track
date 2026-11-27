package services

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Username and password rules (BACKLOG #14).
const (
	PasswordMinLength = 10 // characters
	PasswordMaxBytes  = 72 // bcrypt's input limit is in bytes, not characters
)

var (
	ErrInvalidUsername  = errors.New("username must be 2 to 32 characters: lowercase letters, digits, '.', '_' or '-', starting with a letter or digit")
	ErrPasswordTooShort = errors.New("password must be at least 10 characters")
	ErrPasswordTooLong  = errors.New("password must be at most 72 bytes (fewer characters if it uses accents or emoji)")
)

var usernameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,31}$`)

// NormalizeUsername lowercases and trims a new username and checks its charset. Existing
// usernames aren't re-validated: login matches case-insensitively, whatever they contain.
func NormalizeUsername(username string) (string, error) {
	u := strings.ToLower(strings.TrimSpace(username))
	if !usernameRe.MatchString(u) {
		return "", ErrInvalidUsername
	}
	return u, nil
}

// ValidatePassword checks a password someone is setting (not one they sign in with).
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < PasswordMinLength {
		return ErrPasswordTooShort
	}
	if len(password) > PasswordMaxBytes {
		return ErrPasswordTooLong
	}
	return nil
}
