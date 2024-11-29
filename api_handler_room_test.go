package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestCreateMovie(t *testing.T) {
	token := testLoginAdmin(t)

	input := RoomInput{
		Name: randomString(5),
	}
	newRoom, rec := testCreateRoom(t, token, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)
	require.Equal(t, input.Name, newRoom.Name)
}

func TestCreateRoomFailDuplicate(t *testing.T) {
	token := testLoginAdmin(t)

	input := RoomInput{
		Name: randomString(5),
	}
	newRoom, rec := testCreateRoom(t, token, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)
	require.Equal(t, input.Name, newRoom.Name)

	_, rec = testCreateRoom(t, token, input)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateRoomOK(t *testing.T) {
	token := testLoginAdmin(t)

	newRoom, rec := testCreateRoom(t, token, RoomInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	updated, rec := testUpdateRoom(t, token, newRoom.ID, RoomInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, updated)

	require.Equal(t, newRoom.ID, updated.ID)
	require.NotEqual(t, newRoom.Name, updated.Name)
}

func TestUpdateRoomFailNotFound(t *testing.T) {
	token := testLoginAdmin(t)

	updated, rec := testUpdateRoom(t, token, -21, RoomInput{Name: "not_inserted"})
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, updated)
}

func TestGetRoomOK(t *testing.T) {
	token := testLoginAdmin(t)

	newRoom, rec := testCreateRoom(t, token, RoomInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	cur, rec := testGetRoom(t, newRoom.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, cur)

	require.Equal(t, newRoom.ID, cur.ID)
	require.Equal(t, newRoom.Name, cur.Name)
}

func TestGetRoomFailNotFound(t *testing.T) {
	_, rec := testGetRoom(t, -1)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteRoomOK(t *testing.T) {
	token := testLoginAdmin(t)

	newRoom, rec := testCreateRoom(t, token, RoomInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	_, rec = testDeleteRoom(t, token, newRoom.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	_, rec = testGetRoom(t, newRoom.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPaginateRoomOK(t *testing.T) {
	token := testLoginAdmin(t)

	roomIDs := []int64{}
	for i := 0; i < 5; i++ {
		newRoom, rec := testCreateRoom(t, token, RoomInput{Name: randomString(5)})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, newRoom)
		roomIDs = append(roomIDs, newRoom.ID)
	}

	p, rec := testPaginationRoom(t, RoomFilter{IDs: roomIDs}, PaginateInput{1, 2})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 2)
	require.Equal(t, p.TotalItems, int64(5))

	p, rec = testPaginationRoom(t, RoomFilter{IDs: roomIDs}, PaginateInput{1, 10})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 5)
	require.Equal(t, p.TotalItems, int64(5))
}

func TestSetRoomSeatsOK(t *testing.T) {
	token := testLoginAdmin(t)

	newRoom, rec := testCreateRoom(t, token, RoomInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	//  new seats
	inputSeats := []SeatInput{{Name: randomString(5), AdditionalPrice: 2000}, {Name: randomString(5)}}
	rec = testSetRoomSeats(t, token, newRoom.ID, inputSeats)
	require.Equal(t, http.StatusOK, rec.Code)

	seats, rec := testListRoomSeats(t, newRoom.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, len(inputSeats), len(seats))

	// replace all seats
	inputSeats = []SeatInput{{Name: randomString(5), AdditionalPrice: 2000}, {Name: randomString(5)}, {Name: randomString(3)}}
	rec = testSetRoomSeats(t, token, newRoom.ID, inputSeats)
	require.Equal(t, http.StatusOK, rec.Code)

	seats, rec = testListRoomSeats(t, newRoom.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, len(inputSeats), len(seats))

	// update and append
	inputSeats[1].AdditionalPrice = 15_000
	inputSeats = append(inputSeats, SeatInput{Name: randomString(5)})
	rec = testSetRoomSeats(t, token, newRoom.ID, inputSeats)
	require.Equal(t, http.StatusOK, rec.Code)

	seats, rec = testListRoomSeats(t, newRoom.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, len(inputSeats), len(seats))
}

func testCreateRoom(t *testing.T, token string, input RoomInput) (*Room, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/rooms", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(adminMiddleware(handler.Room.Create))(c)
	require.NoError(t, err)

	var res Response[*Room]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testUpdateRoom(t *testing.T, token string, ID int64, input RoomInput) (*Room, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/rooms/:id", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err = jwtMiddleware(config)(adminMiddleware(handler.Room.UpdateByID))(c)
	require.NoError(t, err)

	var res Response[*Room]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetRoom(t *testing.T, ID int64) (*Room, *httptest.ResponseRecorder) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/rooms/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := handler.Room.GetByID(c)
	require.NoError(t, err)

	var res Response[*Room]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteRoom(t *testing.T, token string, ID int64) (*Room, *httptest.ResponseRecorder) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/rooms/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(adminMiddleware(handler.Room.DeleteByID))(c)
	require.NoError(t, err)

	var res Response[*Room]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPaginationRoom(t *testing.T, filter RoomFilter, page PaginateInput) (*Paginate[Genre], *httptest.ResponseRecorder) {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	e := echo.New()

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))

	req := httptest.NewRequest(http.MethodGet, "/api/rooms?"+q.Encode(), bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.Room.Pagination(c)
	require.NoError(t, err)

	var res Response[*Paginate[Genre]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testSetRoomSeats(t *testing.T, token string, ID int64, input []SeatInput) *httptest.ResponseRecorder {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/:id/seats", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err = jwtMiddleware(config)(adminMiddleware(handler.Room.SetSeats))(c)
	require.NoError(t, err)

	return rec
}

func testListRoomSeats(t *testing.T, ID int64) ([]Seat, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/rooms/:id/seats", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := handler.Room.ListSeats(c)
	require.NoError(t, err)

	var res Response[[]Seat]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
