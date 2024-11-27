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

func (h *MovieHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var input MovieInput
	if err := c.Bind(&input); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok", Data: movie})
}

func (h *MovieHandler) UpdateByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input MovieInput
	if err := c.Bind(&input); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Movie]{Message: "ok", Data: movie})
}

func (h *MovieHandler) DeleteByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Movie.DeleteByID(ctx, int64(ID))
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

func (h *MovieHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Movie]{Message: "ok", Data: movie})
}

func (h *MovieHandler) Pagination(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter MovieFilter
	if err := c.Bind(&filter); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Movie]]{Message: "ok", Data: res})
}

func (h *MovieHandler) CreateGenre(c echo.Context) error {
	ctx := c.Request().Context()

	var input GenreInput
	if err := c.Bind(&input); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Genre]{Message: "ok", Data: genre})
}

func (h *MovieHandler) UpdateGenreByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input GenreInput
	if err := c.Bind(&input); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Genre]{Message: "ok", Data: movie})
}

func (h *MovieHandler) DeleteGenreByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Movie.DeleteGenreByID(ctx, int64(ID))
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

func (h *MovieHandler) PaginationGenre(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter GenreFilter
	if err := c.Bind(&filter); err != nil {
		return err
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
		return err
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Genre]]{Message: "ok", Data: res})
}
