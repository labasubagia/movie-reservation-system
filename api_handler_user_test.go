package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestRole(t *testing.T) {
	token := testLoginAdmin(t)

	pRoles, rec := testPaginationRole(t, token, RoleFilter{Names: []string{"admin"}}, PaginateInput{1, 10})
	require.NotNil(t, pRoles)
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, len(pRoles.Items) > 0)

	role, rec := testGetRole(t, token, pRoles.Items[0].ID)
	require.NotNil(t, role)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "admin", role.Name)
}

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

	cur, rec := testCurrentUser(t, token)
	require.Equal(t, http.StatusOK, rec.Code)
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

func TestChangeUserRoleOK(t *testing.T) {
	token := testLoginAdmin(t)

	input := UserInput{
		Email:    fmt.Sprintf("%s@mail.com", randomString(3)),
		Password: "12345678",
	}
	newUser, rec := testRegisterUser(t, input)
	require.Equal(t, http.StatusOK, rec.Code)

	pRoles, rec := testPaginationRole(t, token, RoleFilter{Names: []string{"admin"}}, PaginateInput{1, 10})
	require.NotNil(t, pRoles)
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, len(pRoles.Items) > 0)
	role := pRoles.Items[0]

	user, rec := testChangeUserRole(t, token, newUser.ID, role.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, user)
	require.Equal(t, user.Role, role.Name)

}

func testRegisterUser(t *testing.T, input UserInput) (*User, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

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

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

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

func testCurrentUser(t *testing.T, token string) (*User, *httptest.ResponseRecorder) {

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*User]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testChangeUserRole(t *testing.T, token string, ID, roleID int64) (*User, *httptest.ResponseRecorder) {
	input := UserInput{RoleID: roleID}
	p, err := json.Marshal(input)
	require.NoError(t, err)

	uri := fmt.Sprintf("/api/admin/user/%d", ID)
	req := httptest.NewRequest(http.MethodPut, uri, bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*User]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetRole(t *testing.T, token string, ID int64) (*Role, *httptest.ResponseRecorder) {
	uri := fmt.Sprintf("/api/admin/roles/%d", ID)
	req := httptest.NewRequest(http.MethodGet, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Role]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPaginationRole(t *testing.T, token string, filter RoleFilter, page PaginateInput) (*Paginate[Role], *httptest.ResponseRecorder) {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))

	uri := "/api/admin/roles?" + q.Encode()
	req := httptest.NewRequest(http.MethodGet, uri, bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Paginate[Role]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
