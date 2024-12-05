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

// Create
//
//	@Summary		Create Room
//	@Description	admin create room
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			request			body		RoomInput	true	"body request"
//	@Success		200				{object}	Response[Room]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/rooms [post]
func (h *RoomHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var input RoomInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

// UpdateByID
//
//	@Summary		Update Room
//	@Description	admin update room by id
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			id				path		int			true	"room id"
//	@Param			request			body		RoomInput	true	"body request"
//	@Success		200				{object}	Response[Room]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/rooms/{id} [put]
func (h *RoomHandler) UpdateByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input RoomInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

// GetByID
//
//	@Summary		Get Room
//	@Description	get room by id
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"room id"
//	@Success		200	{object}	Response[Room]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/rooms/{id} [get]
func (h *RoomHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Room]{Message: "ok", Data: room})
}

// DeleteByID
//
//	@Summary		Delete Room
//	@Description	admin delete room by id
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"room id"
//	@Success		200				{object}	Response[any]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/rooms/{id} [delete]
func (h *RoomHandler) DeleteByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Room.DeleteByID(ctx, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// Pagination
//
//	@Summary		Filter Room
//	@Description	filter rooms
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int			false	"pagination page"
//	@Param			per_page	query		int			false	"pagination page size"
//	@Param			request		body		RoomFilter	false	"filter"
//	@Success		200			{object}	Response[Paginate[Room]]
//	@Failure		400			{object}	Response[any]
//	@Failure		500			{object}	Response[any]
//	@Router			/api/rooms/filter [post]
func (h *RoomHandler) Pagination(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter RoomFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Room]]{Message: "ok", Data: res})
}

// SetSeats
//
//	@Summary		Set room seats
//	@Description	admin set seat for room by room id
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			id				path		int			true	"room id"
//	@Param			request			body		[]SeatInput	true	"body request"
//	@Success		200				{object}	Response[Room]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/rooms/{id}/seats [post]
func (h *RoomHandler) SetSeats(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input []SeatInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// ListSeats
//
//	@Summary		Get room seats
//	@Description	get seats by room id
//	@Tags			rooms
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"room id"
//	@Success		200	{object}	Response[Seat]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/rooms/{id}/seats [get]
func (h *RoomHandler) ListSeats(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var seats []Seat
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		seats, err = service.Room.ListSeats(ctx, int64(ID))
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
