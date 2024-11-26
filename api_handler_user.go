package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func NewUserHandler(c *Config, trxProvider *TransactionProvider) *UserHandler {
	return &UserHandler{
		config:      c,
		trxProvider: trxProvider,
	}
}

type UserHandler struct {
	config      *Config
	trxProvider *TransactionProvider
}

func (h *UserHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var input UserInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, map[string]any{"email": input.Email})

	var user *User
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		user, err = service.User.Register(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: user})
}

func (h *UserHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var input UserInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, map[string]any{"email": input.Email})

	var token string
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		token, err = service.User.Login(ctx, input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: map[string]any{"token": token}})
}

func (h *UserHandler) LoggedIn(c echo.Context) error {
	ctx := c.Request().Context()
	_, email, _ := GetTokenInfo(c)
	c.Set(KeyInput, map[string]any{"email": email})

	var user *User
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		user, err = service.User.GetByEmail(ctx, email)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Response{Message: "ok", Data: user})
}

func (h *UserHandler) ChangeRoleByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewErr(ErrInput, err, "id invalid")
	}

	var input UserInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	c.Set(KeyInput, input)

	var user *User
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		user, err = service.User.ChangeRoleByID(ctx, int64(ID), input)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Response{Message: "ok", Data: user})
}
