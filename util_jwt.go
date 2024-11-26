package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtCustomClaims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func CreateToken(c *Config, user *User) (string, error) {
	claims := &JwtCustomClaims{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(c.JWTSecret))
	if err != nil {
		return "", NewErr(ErrInput, err, "failed to create token, try again later")
	}
	return t, nil
}
