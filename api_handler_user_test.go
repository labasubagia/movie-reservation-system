package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRegisterUserOK(t *testing.T) {
	input := UserInput{
		Email:    "user_for_register@gmail.com",
		Password: "12345678",
	}
	testRegisterUser(t, input)
}

func TestLoginUserOK(t *testing.T) {
	input := UserInput{
		Email:    "user_for_login@gmail.com",
		Password: "12345678",
	}

	newUser := testRegisterUser(t, input)

	token := testLoginUser(t, UserInput{
		Email:    newUser.Email,
		Password: input.Password,
	})

	cur := testCurrentUser(t, token)
	require.Equal(t, newUser.Email, cur.Email)
}

func testRegisterUser(t *testing.T, input UserInput) *User {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.User.Register(c)
	require.NoError(t, err)

	var res Response[*User]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, input.Email, res.Data.Email)

	return res.Data
}

func testLoginUser(t *testing.T, input UserInput) (token string) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.User.Login(c)
	require.NoError(t, err)

	var resLogin Response[map[string]string]
	err = json.Unmarshal(rec.Body.Bytes(), &resLogin)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, resLogin.Data["token"])
	token = fmt.Sprintf("Bearer %s", resLogin.Data["token"])
	return token
}

func testCurrentUser(t *testing.T, token string) *User {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := jwtMiddleware(config)(handler.User.LoggedIn)(c)
	require.NoError(t, err)

	var res Response[*User]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)
	return res.Data
}
