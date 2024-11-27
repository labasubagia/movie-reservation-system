package main

import (
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) PaginateInput {
	p := NewPaginateInput(Page, PageSize)
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err == nil && page > 0 {
		p.Page = int64(page)
	}
	size, err := strconv.Atoi(c.QueryParam("per_page"))
	if err == nil && size > 0 && size <= MaxPageSize {
		p.Size = int64(size)
	}
	return p
}

func GetTokenInfo(c echo.Context) (id int64, email string, role string) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return 0, "", ""
	}
	claims, ok := user.Claims.(*JwtCustomClaims)
	if !ok {
		return 0, "", ""
	}
	return claims.ID, claims.Email, claims.Role
}
