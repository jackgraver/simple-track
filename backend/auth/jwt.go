package auth

import (
	"be-simpletracker/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var secretKey string

func init() {
	utils.LoadEnvIfNeeded()
	secretKey = os.Getenv("JWT_SECRET")
	if secretKey == "" {
		panic("JWT_SECRET environment variable is required but not set")
	}
}

type Claims struct {
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

// GenerateToken creates a JWT token with username and timestamp
func GenerateToken(username string) (string, error) {
	claims := Claims{
		Username:  username,
		Timestamp: time.Now().Unix(),
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signature := createSignature(headerEncoded + "." + claimsEncoded)

	token := fmt.Sprintf("%s.%s.%s", headerEncoded, claimsEncoded, signature)
	return token, nil
}

// VerifyToken validates a JWT token and returns the claims
func VerifyToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	expectedSignature := createSignature(headerEncoded + "." + claimsEncoded)
	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims: %v", err)
	}

	return &claims, nil
}

func createSignature(data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

