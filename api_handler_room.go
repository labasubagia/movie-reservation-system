package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewRoomHandler(c *Config, trxProvider *TransactionProvider) *RoomHandler {
	return &RoomHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type RoomHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

func (h *RoomHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var input RoomInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, input)

	var room *Room
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		room, err = service.Room.Create(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

func (h *RoomHandler) UpdateByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input RoomInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, input)

	var room *Room
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		room, err = service.Room.UpdateByID(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

func (h *RoomHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var room *Room
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		room, err = service.Room.GetByID(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

func (h *RoomHandler) DeleteByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Room.DeleteByID(ctx, int64(ID))
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

func (h *RoomHandler) Pagination(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter RoomFilter
	if err := c.Bind(&filter); err != nil {
		return err
	}
	c.Set(KeyInput, filter)

	var res *Paginate[Room]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Room.Pagination(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Room]]{Message: "ok", Data: res})
}

func (h *RoomHandler) SetSeats(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input []SeatInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, input)

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		err := service.Room.SetSeats(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}
