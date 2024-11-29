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

func TestCreateShowtimeOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newShowtime)
	require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
	require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)
}

func TestCreateShowtimeFailDurationInvalid(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    120,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Minute),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Nil(t, newShowtime)
}

func TestCreateShowtimeFailOverlappingSameRoom(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newShowtime)
	require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
	require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)

	_, rec = testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateShowtimeFailNoRequiredData(t *testing.T) {
	token := testLoginAdmin(t)

	showtimeInput := ShowtimeInput{
		MovieID: -1, // movie not found
		RoomID:  -1,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, newShowtime)
}

func TestUpdateShowtimeOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newShowtime)
	require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
	require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)

	showtimeInput.Price = 60_000
	updatedShowtime, rec := testUpdateShowtime(t, token, newShowtime.ID, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, updatedShowtime)
	require.Equal(t, newShowtime.ID, updatedShowtime.ID)
	require.NotEqual(t, newShowtime.Price, updatedShowtime.Price)
}

func TestUpdateShowtimeFailNotFound(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	updatedShowtime, rec := testUpdateShowtime(t, token, -12, showtimeInput)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, updatedShowtime)
}

func TestGetShowtimeOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newShowtime)
	require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
	require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)

	cur, rec := testGetShowtime(t, newShowtime.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, cur)
	require.Equal(t, newShowtime.ID, cur.ID)
	require.Equal(t, newShowtime.Price, cur.Price)
}

func TestGetShowtimeFailNotFound(t *testing.T) {
	cur, rec := testGetShowtime(t, -12)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, cur)
}

func TestPaginateShowtimeOK(t *testing.T) {

	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	showtimeIDs := []int64{}
	for i := 0; i < 5; i++ {

		newRoom, rec := testCreateRoom(t, token, RoomInput{
			Name: randomString(5),
		})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, newRoom)

		showtimeInput := ShowtimeInput{
			MovieID: newMovie.ID,
			RoomID:  newRoom.ID,
			StartAt: time.Now(),
			EndAt:   time.Now().Add(2 * time.Hour),
			Price:   50_000,
		}

		newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, newShowtime)
		require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
		require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)

		showtimeIDs = append(showtimeIDs, newShowtime.ID)
	}

	p, rec := testPaginateShowtime(t, ShowtimeFilter{IDs: showtimeIDs}, PaginateInput{1, 2})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 2)
	require.Equal(t, p.TotalItems, int64(5))

	p, rec = testPaginateShowtime(t, ShowtimeFilter{IDs: showtimeIDs}, PaginateInput{1, 10})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 5)
	require.Equal(t, p.TotalItems, int64(5))
}

func TestDeleteShowtimeOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(4)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	newMovie, rec := testCreateMovie(t, token, MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    33,
		PosterURL:   fmt.Sprintf("http://%s.com", randomString(5)),
		Description: randomString(5),
		GenreIDs:    []int64{newGenre.ID},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)

	newRoom, rec := testCreateRoom(t, token, RoomInput{
		Name: randomString(5),
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newRoom)

	showtimeInput := ShowtimeInput{
		MovieID: newMovie.ID,
		RoomID:  newRoom.ID,
		StartAt: time.Now(),
		EndAt:   time.Now().Add(2 * time.Hour),
		Price:   50_000,
	}

	newShowtime, rec := testCreateShowtime(t, token, showtimeInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newShowtime)
	require.Equal(t, showtimeInput.MovieID, newShowtime.MovieID)
	require.Equal(t, showtimeInput.RoomID, newShowtime.RoomID)

	_, rec = testDeleteShowtime(t, token, newShowtime.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	_, rec = testGetShowtime(t, newShowtime.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)

}

func testCreateShowtime(t *testing.T, token string, input ShowtimeInput) (*Showtime, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/showtimes", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(adminMiddleware(handler.Showtime.Create))(c)
	require.NoError(t, err)

	var res Response[*Showtime]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testUpdateShowtime(t *testing.T, token string, ID int64, input ShowtimeInput) (*Showtime, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/showtimes/:id", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err = jwtMiddleware(config)(adminMiddleware(handler.Showtime.UpdateByID))(c)
	require.NoError(t, err)

	var res Response[*Showtime]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetShowtime(t *testing.T, ID int64) (*Showtime, *httptest.ResponseRecorder) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/showtimes/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := handler.Showtime.GetByID(c)
	require.NoError(t, err)

	var res Response[*Showtime]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteShowtime(t *testing.T, token string, ID int64) (*Showtime, *httptest.ResponseRecorder) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/showtimes/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(adminMiddleware(handler.Showtime.DeleteByID))(c)
	require.NoError(t, err)

	var res Response[*Showtime]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPaginateShowtime(t *testing.T, filter ShowtimeFilter, page PaginateInput) (*Paginate[Showtime], *httptest.ResponseRecorder) {

	p, err := json.Marshal(filter)
	require.NoError(t, err)

	e := echo.New()

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))

	req := httptest.NewRequest(http.MethodGet, "/api/showtimes?"+q.Encode(), bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.Showtime.Pagination(c)
	require.NoError(t, err)

	var res Response[*Paginate[Showtime]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
