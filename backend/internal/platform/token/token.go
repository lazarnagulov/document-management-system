package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

type Config struct {
	AccessSecret  string
	RefreshSecret string
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAccessToken(userID uuid.UUID, email, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
	return sign(claims, secret)
}

func NewRefreshToken(userID uuid.UUID, secret string) (string, error) {
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
	return sign(claims, secret)
}

func NewRefreshTokenWithDuration(userID uuid.UUID, secret string, d time.Duration) (string, error) {
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
	return sign(claims, secret)
}

func ParseAccessToken(tokenStr, secret string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc(secret))
	if err != nil || !t.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func ParseRefreshToken(tokenStr, secret string) (*RefreshClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, keyFunc(secret))
	if err != nil || !t.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	claims, ok := t.Claims.(*RefreshClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func keyFunc(secret string) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}
}

func sign(claims jwt.Claims, secret string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
