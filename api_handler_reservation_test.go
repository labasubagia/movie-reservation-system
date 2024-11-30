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

func TestReservation(t *testing.T) {
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

	t.Run("ShowtimeIn3Days", func(t *testing.T) {

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

			showStartAt := time.Now().Add(3 * 24 * time.Hour)
			showtime, rec = testCreateShowtime(t, tokenAdmin, ShowtimeInput{
				MovieID: movie.ID,
				RoomID:  room.ID,
				StartAt: showStartAt,
				EndAt:   showStartAt.Add(2 * time.Hour),
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

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
		})

		t.Run("PayOK", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)

			reservation, rec = testPayReservation(t, token, reservation.ID)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationPaid, reservation.Status)
		})

		t.Run("CancelOK", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)

			reservation, rec = testCancelReservation(t, token, reservation.ID)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationCancelled, reservation.Status)
		})

		t.Run("GetOK", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)

			reservation, rec = testGetReservation(t, token, reservation.ID)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)
			require.Len(t, reservation.Items, 2)
		})

		t.Run("DeleteOK", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)

			rec = testDeleteReservation(token, reservation.ID)
			require.Equal(t, http.StatusOK, rec.Code)

			reservation, rec = testGetReservation(t, token, reservation.ID)
			require.Equal(t, http.StatusNotFound, rec.Code)
			require.Nil(t, reservation)
		})

		t.Run("PaginateOK", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			IDs := []int64{}
			for _, seat := range seats {

				cart, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seat.ID})
				require.Equal(t, http.StatusOK, rec.Code)
				require.NotNil(t, cart)

				reservation, rec := testCreateReservation(t, token, ReservationInput{
					CartIDs: []int64{cart.ID},
				})
				require.Equal(t, http.StatusOK, rec.Code)
				require.NotNil(t, reservation)
				require.Equal(t, ReservationUnpaid, reservation.Status)

				IDs = append(IDs, reservation.ID)
			}

			p, rec := testPaginationReservation(t, token, ReservationFilter{IDs: IDs}, PaginateInput{1, 2})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, p)
			require.Len(t, p.Items, 2)
			require.Equal(t, p.TotalItems, int64(5))

			p, rec = testPaginationReservation(t, token, ReservationFilter{IDs: IDs}, PaginateInput{1, 10})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, p)
			require.Len(t, p.Items, 5)
			require.Equal(t, p.TotalItems, int64(5))
		})
	})

	t.Run("ShowtimeNow", func(t *testing.T) {
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

			showStartAt := time.Now()
			showtime, rec = testCreateShowtime(t, tokenAdmin, ShowtimeInput{
				MovieID: movie.ID,
				RoomID:  room.ID,
				StartAt: showStartAt,
				EndAt:   showStartAt.Add(2 * time.Hour),
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

		t.Run("CancelFailed", func(t *testing.T) {
			seats := replaceSeat(room.ID)

			cart1, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[0].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart1)

			cart2, rec := testCreateCart(t, token, CartInput{ShowtimeID: showtime.ID, SeatID: seats[1].ID})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, cart2)

			// create ok
			reservation, rec := testCreateReservation(t, token, ReservationInput{
				CartIDs: []int64{cart1.ID, cart2.ID},
			})
			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, reservation)
			require.Equal(t, ReservationUnpaid, reservation.Status)

			// failed to cancel when movie will play less then 6 hour
			_, rec = testCancelReservation(t, token, reservation.ID)
			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})
}

func testCreateReservation(t *testing.T, token string, input ReservationInput) (*Reservation, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/reservations", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Reservation]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPayReservation(t *testing.T, token string, ID int64) (*Reservation, *httptest.ResponseRecorder) {
	uri := fmt.Sprintf("/api/reservations/%d/pay", ID)
	req := httptest.NewRequest(http.MethodPut, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Reservation]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testCancelReservation(t *testing.T, token string, ID int64) (*Reservation, *httptest.ResponseRecorder) {
	uri := fmt.Sprintf("/api/reservations/%d/cancel", ID)
	req := httptest.NewRequest(http.MethodPut, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Reservation]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetReservation(t *testing.T, token string, ID int64) (*Reservation, *httptest.ResponseRecorder) {
	uri := fmt.Sprintf("/api/reservations/%d", ID)
	req := httptest.NewRequest(http.MethodGet, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Reservation]
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteReservation(token string, ID int64) *httptest.ResponseRecorder {
	uri := fmt.Sprintf("/api/reservations/%d", ID)
	req := httptest.NewRequest(http.MethodDelete, uri, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	return rec
}

func testPaginationReservation(t *testing.T, token string, filter ReservationFilter, page PaginateInput) (*Paginate[Reservation], *httptest.ResponseRecorder) {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))
	uri := "/api/reservations?" + q.Encode()

	req := httptest.NewRequest(http.MethodGet, uri, bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)

	rec := httptest.NewRecorder()
	testServer.ServeHTTP(rec, req)

	var res Response[*Paginate[Reservation]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
