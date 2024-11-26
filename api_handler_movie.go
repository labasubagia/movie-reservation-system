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

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: movie})
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

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: movie})
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

	return c.JSON(http.StatusOK, Response{Message: "ok"})
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

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: res})
}
