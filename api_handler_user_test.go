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
		Email:    fmt.Sprintf("%s@mail.com", randomString(3)),
		Password: "12345678",
	}
	newUser, rec := testRegisterUser(t, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, input.Email, newUser.Email)
}

func TestRegisterUserFailDuplicate(t *testing.T) {
	input := UserInput{
		Email:    fmt.Sprintf("%s@mail.com", randomString(3)),
		Password: "12345678",
	}

	// success
	newUser, rec := testRegisterUser(t, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, input.Email, newUser.Email)

	// fail duplicate
	_, rec = testRegisterUser(t, input)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginUserOK(t *testing.T) {
	input := UserInput{
		Email:    fmt.Sprintf("%s@mail.com", randomString(3)),
		Password: "12345678",
	}

	newUser, rec := testRegisterUser(t, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, input.Email, newUser.Email)

	token, rec := testLoginUser(t, UserInput{
		Email:    newUser.Email,
		Password: input.Password,
	})
	require.NotEqual(t, token, "")
	require.Equal(t, http.StatusOK, rec.Code)

	cur := testCurrentUser(t, token)
	require.Equal(t, newUser.Email, cur.Email)
}

func TestLoginUserFail(t *testing.T) {
	input := UserInput{
		Email:    fmt.Sprintf("%s@mail.com", randomString(3)),
		Password: "12345678",
	}
	_, rec := testLoginUser(t, input)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func testRegisterUser(t *testing.T, input UserInput) (*User, *httptest.ResponseRecorder) {
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

	return res.Data, rec
}

func testLoginAdmin(t *testing.T) (token string) {
	// this data is from migration seed
	input := UserInput{
		Email:    "admin@gmail.com",
		Password: "12345678",
	}
	token, rec := testLoginUser(t, input)
	require.Equal(t, http.StatusOK, rec.Code)

	return token
}

func testLoginUser(t *testing.T, input UserInput) (token string, recorder *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.User.Login(c)
	require.NoError(t, err)

	var res Response[map[string]string]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	_, ok := res.Data["token"]
	if !ok {
		return "", rec
	}
	require.NotEmpty(t, res.Data["token"])
	token = fmt.Sprintf("Bearer %s", res.Data["token"])
	return token, rec
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
