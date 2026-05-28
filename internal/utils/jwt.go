package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type JWTClaims struct {
	UserID     int    `json:"user_id"`
	CustomerID string `json:"customer_id"`
	Role       string `json:"role"`
	ExpiresAt  int64  `json:"exp"`
}

func GenerateJWT(userID int, customerID string, role string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := JWTClaims{
		UserID:     userID,
		CustomerID: customerID,
		Role:       role,
		ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsignedToken := encodedHeader + "." + encodedClaims

	signature := signJWT(unsignedToken)

	return unsignedToken + "." + signature, nil
}

func ParseJWT(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSignature := signJWT(unsignedToken)

	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return nil, errors.New("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	if err := ValidateRole(claims.Role); err != nil {
		return nil, err
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func signJWT(value string) string {
	mac := hmac.New(sha256.New, []byte(jwtSecret()))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func jwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "b1k5-development-secret"
	}

	return secret
}

func ValidateRole(role string) error {
	switch role {
	case "customer", "admin":
		return nil
	default:
		return fmt.Errorf("invalid role: %s", role)
	}
}
