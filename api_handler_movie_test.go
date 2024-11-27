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

func TestCreateMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(5), randomString(5)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre := testCreateGenre(t, token, GenreInput{Name: genre})
		genreIDs = append(genreIDs, genre.ID)
	}

	movieInput := MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    169, // minutes
		PosterURL:   fmt.Sprintf("https://www.%s.com/poster.jpg", randomString(5)),
		Description: randomString(5),
		GenreIDs:    genreIDs,
	}
	newMovie := testCreateMovie(t, token, movieInput)
	require.ElementsMatch(t, genres, newMovie.Genres)
}

func TestUpdateMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4), randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre := testCreateGenre(t, token, GenreInput{Name: genre})
		genreIDs = append(genreIDs, genre.ID)
	}

	movieInput := MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    22, // minutes
		PosterURL:   fmt.Sprintf("https://www.%s.com/poster.jpg", randomString(5)),
		Description: randomString(5),
		GenreIDs:    genreIDs,
	}
	newMovie := testCreateMovie(t, token, movieInput)

	movieInput.Title = randomString(5)
	movieInput.Duration = 148
	movieInput.Director = randomString(5)
	movieInput.GenreIDs = genreIDs[1:]
	updatedMovie := testUpdateMovie(t, token, newMovie.ID, movieInput)

	require.NotEqual(t, newMovie.Duration, updatedMovie.Duration)
	require.NotEqual(t, newMovie.Director, updatedMovie.Director)
	require.NotEqual(t, newMovie.GenreIDs, updatedMovie.GenreIDs)
}

func TestDeleteMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre := testCreateGenre(t, token, GenreInput{Name: genre})
		genreIDs = append(genreIDs, genre.ID)
	}

	movieInput := MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    22, // minutes
		PosterURL:   fmt.Sprintf("https://www.%s.com/poster.jpg", randomString(5)),
		Description: randomString(5),
		GenreIDs:    genreIDs,
	}
	newMovie := testCreateMovie(t, token, movieInput)

	testDeleteMovie(t, token, newMovie.ID)
}

func TestPaginationMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre := testCreateGenre(t, token, GenreInput{Name: genre})
		genreIDs = append(genreIDs, genre.ID)
	}

	movieIDs := []int64{}
	for i := 0; i < 5; i++ {
		movieInput := MovieInput{
			Title:       randomString(5),
			ReleaseDate: time.Now(),
			Director:    randomString(5),
			Duration:    22, // minutes
			PosterURL:   fmt.Sprintf("https://www.%s.com/poster.jpg", randomString(5)),
			Description: randomString(5),
			GenreIDs:    genreIDs,
		}
		v := testCreateMovie(t, token, movieInput)
		movieIDs = append(movieIDs, v.ID)
	}

	p := testPaginationMovie(t, token, MovieFilter{IDs: movieIDs}, PaginateInput{1, 2})
	require.Len(t, p.Items, 2)

	p = testPaginationMovie(t, token, MovieFilter{IDs: movieIDs}, PaginateInput{1, 10})
	require.Len(t, p.Items, 5)
}

func TestGetMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre := testCreateGenre(t, token, GenreInput{Name: genre})
		genreIDs = append(genreIDs, genre.ID)
	}

	movieInput := MovieInput{
		Title:       randomString(5),
		ReleaseDate: time.Now(),
		Director:    randomString(5),
		Duration:    22, // minutes
		PosterURL:   fmt.Sprintf("https://www.%s.com/poster.jpg", randomString(5)),
		Description: randomString(5),
		GenreIDs:    genreIDs,
	}
	newMovie := testCreateMovie(t, token, movieInput)

	curMovie := testGetMovie(t, token, newMovie.ID)
	require.Equal(t, newMovie.ID, curMovie.ID)
	require.Equal(t, newMovie.Title, curMovie.Title)

}

func testCreateGenre(t *testing.T, token string, input GenreInput) *Genre {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler

	req := httptest.NewRequest(http.MethodPost, "/api/admin/genres", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(handler.Movie.CreateGenre)(c)
	require.NoError(t, err)

	var res Response[*Genre]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, res.Data)

	require.Equal(t, input.Name, res.Data.Name)

	return res.Data
}

func testCreateMovie(t *testing.T, token string, input MovieInput) *Movie {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/movies", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(handler.Movie.Create)(c)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, res.Data)

	require.Equal(t, input.Title, res.Data.Title)
	require.Equal(t, input.Director, res.Data.Director)
	require.ElementsMatch(t, input.GenreIDs, res.Data.GenreIDs)

	return res.Data
}

func testUpdateMovie(t *testing.T, token string, ID int64, input MovieInput) *Movie {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/movies/:id", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err = jwtMiddleware(config)(handler.Movie.UpdateByID)(c)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, res.Data)

	require.Equal(t, input.Title, res.Data.Title)
	require.Equal(t, input.Director, res.Data.Director)
	require.ElementsMatch(t, input.GenreIDs, res.Data.GenreIDs)

	return res.Data
}

func testDeleteMovie(t *testing.T, token string, ID int64) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/movies/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(handler.Movie.DeleteByID)(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

func testGetMovie(t *testing.T, token string, ID int64) *Movie {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/movies/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(handler.Movie.GetByID)(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.NotNil(t, res.Data)

	return res.Data
}

func testPaginationMovie(t *testing.T, token string, filter MovieFilter, page PaginateInput) *Paginate[Movie] {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	e := echo.New()

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/movies?"+q.Encode(), bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(handler.Movie.Pagination)(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var res Response[*Paginate[Movie]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	require.NotNil(t, res.Data)

	return res.Data
}
