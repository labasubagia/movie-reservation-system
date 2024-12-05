package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewMovieHandler(c *Config, trxProvider *TransactionProvider) *MovieHandler {
	return &MovieHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type MovieHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

// Create
//
//	@Summary		Create Movie
//	@Description	admin create movie
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			request			body		MovieInput	true	"body request"
//	@Success		200				{object}	Response[Movie]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/movies [post]
func (h *MovieHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var input MovieInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var movie *Movie
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		movie, err = service.Movie.Create(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok", Data: movie})
}

// UpdateByID
//
//	@Summary		Update Movie
//	@Description	admin update movie by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			id				path		int			true	"movie id"
//	@Param			request			body		MovieInput	true	"body request"
//	@Success		200				{object}	Response[Movie]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/movies/{id} [put]
func (h *MovieHandler) UpdateByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input MovieInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var movie *Movie
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		movie, err = service.Movie.UpdateByID(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Movie]{Message: "ok", Data: movie})
}

// DeleteByID
//
//	@Summary		Delete Movie
//	@Description	admin delete movie by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"movie id"
//	@Success		200				{object}	Response[any]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/movies/{id} [delete]
func (h *MovieHandler) DeleteByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Movie.DeleteByID(ctx, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// GetByID
//
//	@Summary		Get Movie
//	@Description	get movie by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"movie id"
//	@Success		200	{object}	Response[Movie]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/movies/{id} [get]
func (h *MovieHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var movie *Movie
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		movie, err = service.Movie.GetByID(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Movie]{Message: "ok", Data: movie})
}

// Pagination
//
//	@Summary		Filter Movie
//	@Description	filter movies
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int			false	"pagination page"
//	@Param			per_page	query		int			false	"pagination page size"
//	@Param			request		body		MovieFilter	false	"filter"
//	@Success		200			{object}	Response[Paginate[Movie]]
//	@Failure		400			{object}	Response[any]
//	@Failure		500			{object}	Response[any]
//	@Router			/api/movies/filter [post]
func (h *MovieHandler) Pagination(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter MovieFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, filter)

	var res *Paginate[Movie]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Movie.Pagination(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Movie]]{Message: "ok", Data: res})
}

// CreateGenre
//
//	@Summary		Create Genre
//	@Description	admin create genre
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			request			body		GenreInput	true	"body request"
//	@Success		200				{object}	Response[Genre]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/genres [post]
func (h *MovieHandler) CreateGenre(c echo.Context) error {
	ctx := c.Request().Context()

	var input GenreInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var genre *Genre
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		genre, err = service.Movie.CreateGenre(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Genre]{Message: "ok", Data: genre})
}

// UpdateGenreByID
//
//	@Summary		Update Genre
//	@Description	admin update genre by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			id				path		int			true	"genre id"
//	@Param			request			body		GenreInput	true	"body request"
//	@Success		200				{object}	Response[Genre]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/genres/{id} [put]
func (h *MovieHandler) UpdateGenreByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input GenreInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var movie *Genre
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		movie, err = service.Movie.UpdateGenreByID(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Genre]{Message: "ok", Data: movie})
}

// DeleteGenreByID
//
//	@Summary		Delete Genre
//	@Description	admin delete delete by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"genre id"
//	@Success		200				{object}	Response[any]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/genres/{id} [delete]
func (h *MovieHandler) DeleteGenreByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Movie.DeleteGenreByID(ctx, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// GetGenreByID
//
//	@Summary		Get Genre
//	@Description	get genre by id
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"genre id"
//	@Success		200	{object}	Response[any]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/genres/{id} [get]
func (h *MovieHandler) GetGenreByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var genre *Genre
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		genre, err = service.Movie.GetGenreByID(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Genre]{Message: "ok", Data: genre})
}

// PaginationGenre
//
//	@Summary		Filter Genre
//	@Description	filter genre
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int			false	"pagination page"
//	@Param			per_page	query		int			false	"pagination page size"
//	@Param			request		body		GenreFilter	false	"filter"
//	@Success		200			{object}	Response[Paginate[Genre]]
//	@Failure		400			{object}	Response[any]
//	@Failure		500			{object}	Response[any]
//	@Router			/api/genres/filter [post]
func (h *MovieHandler) PaginationGenre(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter GenreFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, filter)

	var res *Paginate[Genre]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Movie.PaginationGenre(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Genre]]{Message: "ok", Data: res})
}
