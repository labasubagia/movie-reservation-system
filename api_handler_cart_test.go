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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestCart(t *testing.T) {

	var showtime *Showtime
	var room *Room

	// admin space
	tokenAdmin := testLoginAdmin(t)

	replaceSeat := func(roomID int64) [5]Seat {

		inputSeats := []SeatInput{
			{Name: randomString(5), AdditionalPrice: 3000},
			{Name: randomString(5), AdditionalPrice: 1000},
			{Name: randomString(5), AdditionalPrice: 0},
			{Name: randomString(5), AdditionalPrice: 500},
			{Name: randomString(5), AdditionalPrice: 2000},
		}
		rec := testSetRoomSeats(t, tokenAdmin, roomID, inputSeats)
		require.Equal(t, http.StatusOK, rec.Code)

		seats, rec := testListRoomSeats(t, roomID)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, len(inputSeats), len(seats))

		return [5]Seat(seats)
	}

	// admin auth
	{
		genre, rec := testCreateGenre(t, tokenAdmin, GenreInput{Name: randomString(4)})
		require.Equal(t, http.StatusOK, rec.Code)

		movie, rec := testCreateMovie(t, tokenAdmin, MovieInput{
			Title:       randomString(5),
			ReleaseDate: time.Now(),
			Director:    randomString(5),
			Duration:    33,
			PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
			Description: randomString(5),
			GenreIDs:    []int64{genre.ID},
		})
		require.Equal(t, http.StatusOK, rec.Code)

		room, rec = testCreateRoom(t, tokenAdmin, RoomInput{
			Name: randomString(5),
		})
		require.Equal(t, http.StatusOK, rec.Code)

		showtime, rec = testCreateShowtime(t, tokenAdmin, ShowtimeInput{
			MovieID: movie.ID,
			RoomID:  room.ID,
			StartAt: time.Now(),
			EndAt:   time.Now().Add(movie.GetDuration()),
			Price:   50_000,
		})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, showtime)
	}

	userInput := UserInput{
		Email:    fmt.Sprintf("%s@gmail.com", randomString(5)),
		Password: "12345678",
	}
	user, rec := testRegisterUser(t, userInput)
	require.Equal(t, http.StatusOK, rec.Code)

	token, rec := testLoginUser(t, UserInput{Email: user.Email, Password: userInput.Password})

	t.Run("CreateOK", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)
	})

	t.Run("CreateFailDuplicate", func(t *testing.T) {

		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)

		_, rec = testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("UpdateOK", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)

		_, rec = testUpdateCart(t, token, cart.ID, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("UpdateFailInvalidShowtime", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)

		_, rec = testUpdateCart(t, token, cart.ID, CartInput{ShowtimeID: -1, SeatID: seats[0].ID})
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("GetOK", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)

		cur, rec := testGetCart(t, token, cart.ID)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cur)

		require.Equal(t, cart.ID, cur.ID)
		require.Equal(t, cart.ShowtimeID, cur.ShowtimeID)
		require.Equal(t, cart.Movie, cur.Movie)

	})

	t.Run("GetFailNotFound", func(t *testing.T) {

		cur, rec := testGetCart(t, token, -1)
		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Nil(t, cur)
	})

	t.Run("DeleteOK", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, cart)

		rec = testDeleteCart(token, cart.ID)
		require.Equal(t, http.StatusOK, rec.Code)

		_, rec = testGetCart(t, token, cart.ID)
		require.Equal(t, http.StatusNotFound, rec.Code)

	})

	t.Run("PaginationOK", func(t *testing.T) {
		seats := replaceSeat(room.ID)

		cartIDs := []int64{}
		for i := 0; i < 5; i++ {
			cart, rec := testCreateCart(t, token, CartInput{
				SeatID:     seats[i].ID,
				ShowtimeID: showtime.ID,
			})

			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart)
			cartIDs = append(cartIDs, cart.ID)
		}

		p, rec := testPaginationCart(t, token, RoomFilter{IDs: cartIDs}, PaginateInput{1, 2})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, p)
		require.Len(t, p.Items, 2)
		require.Equal(t, p.TotalItems, int64(5))

		p, rec = testPaginationCart(t, token, RoomFilter{IDs: cartIDs}, PaginateInput{1, 10})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, p)
		require.Len(t, p.Items, 5)
		require.Equal(t, p.TotalItems, int64(5))
	})

}

func testCreateCart(t *testing.T, token string, input CartInput) (*Cart, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/carts", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Cart]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testUpdateCart(t *testing.T, token string, ID int64, input CartInput) (*Cart, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	uri := fmt.Sprintf("/api/carts/%d", ID)
	req := httptest.NewRequest(http.MethodPut, uri, bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Cart]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetCart(t *testing.T, token string, ID int64) (*Cart, *httptest.ResponseRecorder) {
	uri := fmt.Sprintf("/api/carts/%d", ID)
	req := httptest.NewRequest(http.MethodGet, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Cart]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteCart(token string, ID int64) *httptest.ResponseRecorder {
	uri := fmt.Sprintf("/api/carts/%d", ID)
	req := httptest.NewRequest(http.MethodDelete, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	return rec
}

func testPaginationCart(t *testing.T, token string, filter RoomFilter, page PaginateInput) (*Paginate[Cart], *httptest.ResponseRecorder) {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))
	uri := "/api/carts?" + q.Encode()

	req := httptest.NewRequest(http.MethodGet, uri, bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)

	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Paginate[Cart]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
