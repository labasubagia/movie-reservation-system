package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewCartHandler(c *Config, trxProvider *TransactionProvider) *CartHandler {
	return &CartHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type CartHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

// UserCreate
//
//	@Summary		Create Cart
//	@Description	user create cart showtime
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			request			body		CartInput	true	"body request"
//	@Success		200				{object}	Response[Cart]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/carts [post]
func (h *CartHandler) UserCreate(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	var input CartInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	input.UserID = userID
	c.Set(KeyInput, input)

	var cart *Cart
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		cart, err = service.Cart.Create(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Cart]{Message: "ok", Data: cart})
}

// UserUpdateByID
//
//	@Summary		Update Cart
//	@Description	user update cart by id
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			id				path		int			true	"cart id"
//	@Param			request			body		CartInput	true	"body request"
//	@Success		200				{object}	Response[Cart]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/carts/{id} [put]
func (h *CartHandler) UserUpdateByID(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input CartInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	input.UserID = userID

	c.Set(KeyInput, input)

	var cart *Cart
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		cart, err = service.Cart.UserUpdateByID(ctx, userID, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Cart]{Message: "ok", Data: cart})
}

// UserGetByID
//
//	@Summary		Get Cart
//	@Description	user get cart by cart id
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"cart id"
//	@Success		200				{object}	Response[Cart]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/carts/{id} [get]
func (h *CartHandler) UserGetByID(c echo.Context) error {
	userID, _, _ := GetTokenInfo(c)

	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var cart *Cart
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		cart, err = service.Cart.UserGetByID(ctx, userID, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Cart]{Message: "ok", Data: cart})
}

// UserDeleteByID
//
//	@Summary		Delete Cart
//	@Description	user delete cart by cart by id
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"bearer token"
//	@Param			id				path		int		true	"cart id"
//	@Success		200				{object}	Response[any]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/carts/{id} [delete]
func (h *CartHandler) UserDeleteByID(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		return service.Cart.UserDeleteByID(ctx, userID, int64(ID))
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok"})
}

// UserGetPagination
//
//	@Summary		Filter Cart
//	@Description	user filter own carts
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			page			query		int			false	"pagination page"
//	@Param			per_page		query		int			false	"pagination page size"
//	@Param			request			body		CartFilter	false	"filter"
//	@Success		200				{object}	Response[Paginate[Cart]]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/carts/filter [post]
func (h *CartHandler) UserGetPagination(c echo.Context) error {
	ctx := c.Request().Context()
	userID, _, _ := GetTokenInfo(c)

	page := GetPage(c)

	var filter CartFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
	}
	filter.UserIDs = []int64{userID}
	c.Set(KeyInput, filter)

	var res *Paginate[Cart]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.Cart.Pagination(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Cart]]{Message: "ok", Data: res})
}
