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

func TestCreateGenreOK(t *testing.T) {
	token := testLoginAdmin(t)
	testCreateGenre(t, token, GenreInput{Name: randomString(5)})
}

func TestCreateGenreFailDuplicate(t *testing.T) {
	token := testLoginAdmin(t)
	input := GenreInput{Name: randomString(5)}

	// success
	testCreateGenre(t, token, input)

	// fail duplicate
	_, rec := testCreateGenre(t, token, input)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateGenreOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	input := GenreInput{randomString(4)}

	updated1, rec := testUpdateGenre(t, token, newGenre.ID, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, updated1)
	require.Equal(t, newGenre.ID, updated1.ID)
	require.NotEqual(t, newGenre.Name, updated1.Name)

	updated2, rec := testUpdateGenre(t, token, newGenre.ID, input)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, updated1)
	require.Equal(t, updated1.ID, updated2.ID)
	require.Equal(t, updated1.Name, updated2.Name)
}

func TestUpdateGenreFailNotFound(t *testing.T) {
	token := testLoginAdmin(t)

	updated, rec := testUpdateGenre(t, token, -12, GenreInput{Name: "not_executed"})
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, updated)
}

func TestGetGenreOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	cur, rec := testGetGenre(t, token, newGenre.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, cur)
	require.Equal(t, newGenre.Name, cur.Name)
}

func TestGetGenreFailNotFound(t *testing.T) {
	token := testLoginAdmin(t)

	cur, rec := testGetGenre(t, token, -12)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Nil(t, cur)
}

func TestDeleteGenreOK(t *testing.T) {
	token := testLoginAdmin(t)
	newGenre, rec := testCreateGenre(t, token, GenreInput{Name: randomString(5)})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newGenre)

	cur, rec := testDeleteGenre(t, token, newGenre.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Nil(t, cur)

	cur, rec = testDeleteGenre(t, token, newGenre.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Nil(t, cur)
}

func TestPaginationGenreOK(t *testing.T) {
	token := testLoginAdmin(t)

	genreIDs := []int64{}
	for i := 0; i < 5; i++ {
		input := GenreInput{
			Name: randomString(5),
		}
		genre, rec := testCreateGenre(t, token, input)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

		genreIDs = append(genreIDs, genre.ID)
	}

	p, rec := testPaginationGenre(t, token, GenreFilter{IDs: genreIDs}, PaginateInput{1, 2})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 2)

	p, rec = testPaginationGenre(t, token, GenreFilter{IDs: genreIDs}, PaginateInput{1, 10})
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, p)
	require.Len(t, p.Items, 5)
}

func TestCreateMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(5), randomString(5)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre, rec := testCreateGenre(t, token, GenreInput{Name: genre})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

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
	newMovie, rec := testCreateMovie(t, token, movieInput)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)
	require.Equal(t, movieInput.Title, newMovie.Title)
	require.Equal(t, movieInput.Director, newMovie.Director)
	require.ElementsMatch(t, movieInput.GenreIDs, newMovie.GenreIDs)
	require.ElementsMatch(t, genres, newMovie.Genres)
}

func TestUpdateMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4), randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre, rec := testCreateGenre(t, token, GenreInput{Name: genre})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

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
	newMovie, rec := testCreateMovie(t, token, movieInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)
	require.Equal(t, movieInput.Title, newMovie.Title)
	require.Equal(t, movieInput.Director, newMovie.Director)
	require.ElementsMatch(t, movieInput.GenreIDs, newMovie.GenreIDs)
	require.ElementsMatch(t, genres, newMovie.Genres)

	movieInput.Title = randomString(5)
	movieInput.Duration = 148
	movieInput.Director = randomString(5)
	movieInput.GenreIDs = genreIDs[1:]
	updatedMovie, rec := testUpdateMovie(t, token, newMovie.ID, movieInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, updatedMovie)

	require.NotEqual(t, newMovie.Duration, updatedMovie.Duration)
	require.NotEqual(t, newMovie.Director, updatedMovie.Director)
	require.NotEqual(t, newMovie.GenreIDs, updatedMovie.GenreIDs)
}

func TestDeleteMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre, rec := testCreateGenre(t, token, GenreInput{Name: genre})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

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
	newMovie, rec := testCreateMovie(t, token, movieInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)
	require.Equal(t, movieInput.Title, newMovie.Title)
	require.Equal(t, movieInput.Director, newMovie.Director)
	require.ElementsMatch(t, movieInput.GenreIDs, newMovie.GenreIDs)
	require.ElementsMatch(t, genres, newMovie.Genres)

	rec = testDeleteMovie(t, token, newMovie.ID)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPaginationMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre, rec := testCreateGenre(t, token, GenreInput{Name: genre})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

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
		newMovie, rec := testCreateMovie(t, token, movieInput)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, newMovie)
		movieIDs = append(movieIDs, newMovie.ID)
	}

	p, rec := testPaginationMovie(t, token, MovieFilter{IDs: movieIDs}, PaginateInput{1, 2})
	require.NotNil(t, p)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, p.Items, 2)

	p, rec = testPaginationMovie(t, token, MovieFilter{IDs: movieIDs}, PaginateInput{1, 10})
	require.NotNil(t, p)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, p.Items, 5)
}

func TestGetMovieOK(t *testing.T) {
	token := testLoginAdmin(t)

	genres := []string{randomString(4)}
	genreIDs := []int64{}
	for _, genre := range genres {
		genre, rec := testCreateGenre(t, token, GenreInput{Name: genre})
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, genre)

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
	newMovie, rec := testCreateMovie(t, token, movieInput)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, newMovie)
	require.Equal(t, movieInput.Title, newMovie.Title)
	require.Equal(t, movieInput.Director, newMovie.Director)
	require.ElementsMatch(t, movieInput.GenreIDs, newMovie.GenreIDs)
	require.ElementsMatch(t, genres, newMovie.Genres)

	curMovie, rec := testGetMovie(t, token, newMovie.ID)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, curMovie)

	require.Equal(t, newMovie.ID, curMovie.ID)
	require.Equal(t, newMovie.Title, curMovie.Title)

}

func testCreateGenre(t *testing.T, token string, input GenreInput) (*Genre, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/genres", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(adminMiddleware(handler.Movie.CreateGenre))(c)
	require.NoError(t, err)

	var res Response[*Genre]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testUpdateGenre(t *testing.T, token string, ID int64, input GenreInput) (*Genre, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/genres/:id", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err = jwtMiddleware(config)(adminMiddleware(handler.Movie.UpdateGenreByID))(c)
	require.NoError(t, err)

	var res Response[*Genre]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testGetGenre(t *testing.T, token string, ID int64) (*Genre, *httptest.ResponseRecorder) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/genres/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(adminMiddleware(handler.Movie.GetGenreByID))(c)
	require.NoError(t, err)

	var res Response[*Genre]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteGenre(t *testing.T, token string, ID int64) (*Genre, *httptest.ResponseRecorder) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/genres/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(adminMiddleware(handler.Movie.DeleteGenreByID))(c)
	require.NoError(t, err)

	var res Response[*Genre]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPaginationGenre(t *testing.T, token string, filter GenreFilter, page PaginateInput) (*Paginate[Genre], *httptest.ResponseRecorder) {
	p, err := json.Marshal(filter)
	require.NoError(t, err)

	e := echo.New()

	q := make(url.Values)
	q.Set("page", strconv.Itoa(int(page.Page)))
	q.Set("per_page", strconv.Itoa(int(page.Size)))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/genres?"+q.Encode(), bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(adminMiddleware(handler.Movie.PaginationGenre))(c)
	require.NoError(t, err)

	var res Response[*Paginate[Genre]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testCreateMovie(t *testing.T, token string, input MovieInput) (*Movie, *httptest.ResponseRecorder) {
	p, err := json.Marshal(input)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/movies", bytes.NewReader(p))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = jwtMiddleware(config)(adminMiddleware(handler.Movie.Create))(c)
	require.NoError(t, err)

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testUpdateMovie(t *testing.T, token string, ID int64, input MovieInput) (*Movie, *httptest.ResponseRecorder) {
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

	err = jwtMiddleware(config)(adminMiddleware(handler.Movie.UpdateByID))(c)
	require.NoError(t, err)

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testDeleteMovie(t *testing.T, token string, ID int64) *httptest.ResponseRecorder {

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/movies/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(ID)))

	err := jwtMiddleware(config)(adminMiddleware(handler.Movie.DeleteByID))(c)
	require.NoError(t, err)
	return rec
}

func testGetMovie(t *testing.T, token string, ID int64) (*Movie, *httptest.ResponseRecorder) {

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

	var res Response[*Movie]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}

func testPaginationMovie(t *testing.T, token string, filter MovieFilter, page PaginateInput) (*Paginate[Movie], *httptest.ResponseRecorder) {
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

	var res Response[*Paginate[Movie]]
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	return res.Data, rec
}
