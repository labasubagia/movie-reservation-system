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
