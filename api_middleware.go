package main

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echo_jwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func jwtMiddleware(config *Config) echo.MiddlewareFunc {
	return echo_jwt.WithConfig(echo_jwt.Config{
		SigningKey: []byte(config.JWTSecret),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JwtCustomClaims)
		},
	})
}

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, _, role := GetTokenInfo(c)
		if role == UserAdmin {
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, Response[any]{Message: "unauthorized"})
	}
}
