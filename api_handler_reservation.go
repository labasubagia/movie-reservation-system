package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewReservationHandler(c *Config, trxProvider *TransactionProvider) *ReservationHandler {
	return &ReservationHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type ReservationHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

func (h *ReservationHandler) UserCreate(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	var input ReservationInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	input.UserID = userID
	c.Set(KeyInput, input)

	var reservation *Reservation
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.Create(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

func (h *ReservationHandler) UserUpdateByID(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input ReservationInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	input.UserID = userID

	c.Set(KeyInput, input)

	var reservation *Reservation
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.UserUpdateByID(ctx, userID, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

func (h *ReservationHandler) UserGetByID(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var reservation *Reservation
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.UserGetByID(ctx, userID, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

func (h *ReservationHandler) UserDeleteByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Reservation.UserDeleteByID(ctx, userID, int64(ID))
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

func (h *ReservationHandler) UserGetPagination(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	page := GetPage(c)

	var filter ReservationFilter
	if err := c.Bind(&filter); err != nil {
		return err
	}
	filter.UserIDs = []int64{userID}
	c.Set(KeyInput, filter)

	var res *Paginate[Reservation]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Reservation.Pagination(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Reservation]]{Message: "ok", Data: res})
}
