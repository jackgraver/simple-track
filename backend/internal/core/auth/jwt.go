package auth

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/env"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidToken wraps every reason a token is rejected (bad format, signature, expired,
// revoked...). Errors that don't wrap it are server-side failures (missing secret, DB down).
var ErrInvalidToken = errors.New("invalid token")

func invalidToken(reason string) error { return fmt.Errorf("%w: %s", ErrInvalidToken, reason) }

// Claims is the JWT payload. UID and Ver identify the session: the middleware rejects a token
// without a UID (issued before multi-user) and one whose Ver no longer matches the user's
// token_version (revoked by a password reset or disable).
type Claims struct {
	UID       uint   `json:"uid,omitempty"`
	Ver       int64  `json:"ver,omitempty"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
	Iat       int64  `json:"iat,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
}

// GenerateToken signs a session token for user. The user must be a stored row (non-zero ID).
func GenerateToken(user models.User) (string, error) {
	if user.ID == 0 {
		return "", fmt.Errorf("cannot issue a token for an unsaved user")
	}
	now := clock.Now().Unix()
	ttl := int64(CookieMaxAgeSeconds())
	claims := Claims{
		UID:       user.ID,
		Ver:       user.TokenVersion,
		Username:  user.Username,
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
		return nil, invalidToken("format")
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signature := parts[2]

	expectedSignature, err := createSignature(headerEncoded + "." + claimsEncoded)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, invalidToken("signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, invalidToken("encoding: " + err.Error())
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, invalidToken("claims: " + err.Error())
	}

	if claims.Exp > 0 && clock.Now().Unix() > claims.Exp {
		return nil, invalidToken("expired")
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
