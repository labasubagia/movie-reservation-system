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
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*User]{Message: "ok", Data: user})
}

func (h *UserHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var input UserInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[any]{Message: "ok", Data: map[string]string{"token": token}})
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
		return NewAPIErr(c, err)
	}
	return c.JSON(http.StatusOK, Response[*User]{Message: "ok", Data: user})
}

func (h *UserHandler) ChangeRoleByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input UserInput
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
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
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*User]{Message: "ok", Data: user})
}

func (h *UserHandler) GetRoleByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var role *Role
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		role, err = service.User.GetRoleByID(ctx, int64(ID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Role]{Message: "ok", Data: role})
}

func (h *UserHandler) PaginationRole(c echo.Context) error {
	ctx := c.Request().Context()
	page := GetPage(c)

	var filter RoleFilter
	if err := c.Bind(&filter); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, filter)

	var res *Paginate[Role]
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		res, err = service.User.PaginationRole(ctx, filter, page)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return NewAPIErr(c, err)
	}

	return c.JSON(http.StatusOK, Response[*Paginate[Role]]{Message: "ok", Data: res})
}
