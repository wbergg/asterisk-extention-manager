package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID        int    `json:"user_id"`
	Username      string `json:"username"`
	Role          string `json:"role"`
	MinExt        int    `json:"min_ext"`
	MaxExt        int    `json:"max_ext"`
	CallLogAccess bool   `json:"call_log_access"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, userID int, username, role string, minExt, maxExt int, callLogAccess bool) (string, error) {
	claims := Claims{
		UserID:        userID,
		Username:      username,
		Role:          role,
		MinExt:        minExt,
		MaxExt:        maxExt,
		CallLogAccess: callLogAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*Claims), nil
}
