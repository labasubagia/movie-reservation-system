package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewShowtimeHandler(c *Config, trxProvider *TransactionProvider) *ShowtimeHandler {
	return &ShowtimeHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type ShowtimeHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

func (h *ShowtimeHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var input ShowtimeInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var showtime *Showtime
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		showtime, err = service.Showtime.Create(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Showtime]{Message: "ok", Data: showtime})
}

func (h *ShowtimeHandler) UpdateByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input ShowtimeInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var showtime *Showtime
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		showtime, err = service.Showtime.UpdateByID(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Showtime]{Message: "ok", Data: showtime})
}

func (h *ShowtimeHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var showtime *Showtime
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		showtime, err = service.Showtime.GetByID(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Showtime]{Message: "ok", Data: showtime})
}

func (h *ShowtimeHandler) DeleteByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Showtime.DeleteByID(ctx, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Showtime]{Message: "ok"})
}

func (h *ShowtimeHandler) Pagination(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter ShowtimeFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, filter)

	var res *Paginate[Showtime]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Showtime.Pagination(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Showtime]]{Message: "ok", Data: res})
}

func (h *ShowtimeHandler) GetShowtimeSeatByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var seats []Seat
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		seats, err = service.Showtime.GetShowtimeSeats(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[[]Seat]{Message: "ok", Data: seats})
}
