package auth

import (
	"be-simpletracker/internal/env"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
	Iat       int64  `json:"iat,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
}

func GenerateToken(username string) (string, error) {
	now := time.Now().Unix()
	ttl := int64(CookieMaxAgeSeconds())
	claims := Claims{
		Username:  username,
		Timestamp: now,
		Iat:       now,
		Exp:       now + ttl,
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

	signature, err := createSignature(headerEncoded + "." + claimsEncoded)
	if err != nil {
		return "", err
	}

	token := fmt.Sprintf("%s.%s.%s", headerEncoded, claimsEncoded, signature)
	return token, nil
}

func VerifyToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	expectedSignature, err := createSignature(headerEncoded + "." + claimsEncoded)
	if err != nil {
		return nil, err
	}
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

	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func createSignature(data string) (string, error) {
	secret, err := env.String("JWT_SECRET")
	
	if err != nil {
		return "", err
	}
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))

	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}