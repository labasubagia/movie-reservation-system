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

type ReservationUserCreateReq struct {
	CartIDs []int64 `json:"cart_ids"`
}

// UserCreate
//
//	@Summary		Create Reservation
//	@Description	user create reservations
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"bearer token"
//	@Param			request			body		ReservationUserCreateReq	true	"body request"
//	@Success		200				{object}	Response[Reservation]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations [post]
func (h *ReservationHandler) UserCreate(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	var input ReservationUserCreateReq
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var reservation *Reservation
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.Create(ctx, ReservationInput{
			UserID:  userID,
			CartIDs: input.CartIDs,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

// Pay
//
//	@Summary		Pay Reservation
//	@Description	user pay reservation by id
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"reservation id"
//	@Success		200				{object}	Response[Reservation]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations/{id}/pay [put]
func (h *ReservationHandler) Pay(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var reservation *Reservation
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.Pay(ctx, userID, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

// Cancel
//
//	@Summary		Cancel Reservation
//	@Description	user cancel reservation by id
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"reservation id"
//	@Success		200				{object}	Response[Reservation]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations/{id}/cancel [put]
func (h *ReservationHandler) Cancel(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var reservation *Reservation
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		reservation, err = service.Reservation.Cancel(ctx, userID, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

// UserUpdateByID
//
//	@Summary		Update Reservation
//	@Description	user update reservation by id
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"reservation id"
//	@Success		200				{object}	Response[Reservation]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations/{id} [get]
func (h *ReservationHandler) UserGetByID(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Reservation]{Message: "ok", Data: reservation})
}

// UserDeleteByID
//
//	@Summary		Delete Reservation
//	@Description	user delete reservation by reservation by id
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"reservation id"
//	@Success		200				{object}	Response[any]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations/{id} [delete]
func (h *ReservationHandler) UserDeleteByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Reservation.UserDeleteByID(ctx, userID, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// UserGetPagination
//
//	@Summary		Filter Reservation
//	@Description	user filter own reservations
//	@Tags			reservations
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"bearer token"
//	@Param			page			query		int					false	"pagination page"
//	@Param			per_page		query		int					false	"pagination page size"
//	@Param			request			body		ReservationFilter	false	"filter"
//	@Success		200				{object}	Response[Paginate[Reservation]]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/reservations/filter [post]
func (h *ReservationHandler) UserGetPagination(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	page := GetPage(c)

	var filter ReservationFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Reservation]]{Message: "ok", Data: res})
}
