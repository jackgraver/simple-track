package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

func Load() error {
	loadOnce.Do(func() {
		if err := godotenv.Load(".env"); err != nil {
			_ = godotenv.Load("../../.env")
		}
	})
	return nil
}

var ErrMissing = errors.New("required env var not set")

func String(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("%w: %s", ErrMissing, key)
	}
	return v, nil
}

func StringOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func OptionalString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func Int(key string) (int, error) {
	s, err := String(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid int for %s: %q", key, s)
	}
	return v, nil
}

func IntOr(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func splitTrim(key, sep string) []string {
	raw := strings.Split(os.Getenv(key), sep)
	var out []string
	for _, p := range raw {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func Slice(key, sep string) ([]string, error) {
	out := splitTrim(key, sep)
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %s (expected non-empty list)", ErrMissing, key)
	}
	return out, nil
}

func SliceOr(key, sep string, fallback []string) []string {
	out := splitTrim(key, sep)
	if len(out) == 0 {
		return append([]string(nil), fallback...)
	}
	return out
}
