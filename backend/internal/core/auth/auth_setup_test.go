package auth

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-jwt-secret-32-characters!!")
	os.Exit(m.Run())
}
