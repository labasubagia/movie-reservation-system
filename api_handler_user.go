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

type RegisterUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register
//
//	@Summary		Register New User
//	@Description	register using email and password
//	@Tags			accounts
//	@Accept			json
//	@Param			request	body	RegisterUserReq	true	"req"
//	@Produce		json
//	@Success		200	{object}	Response[User]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/register [post]
func (h *UserHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var input RegisterUserReq
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, map[string]any{"email": input.Email})

	var user *User
	var err error
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		user, err = service.User.Register(ctx, UserInput{Email: input.Email, Password: input.Password})
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

type LoginUserReq = RegisterUserReq

type LoginUserRes struct {
	Token string `json:"token"`
}

// Login
//
//	@Summary		Login User
//	@Description	login using email and password
//	@Tags			accounts
//	@Accept			json
//	@Param			request	body	LoginUserReq	true	"req"
//	@Produce		json
//	@Success		200	{object}	Response[LoginUserRes]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/login [post]
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

	return c.JSON(http.StatusOK, Response[any]{Message: "ok", Data: LoginUserRes{Token: token}})
}

// LoggedIn get current user using token
//
//	@Summary		Current User
//	@Description	get information of current user
//	@Tags			accounts
//	@Accept			json
//	@Param			Authorization	header	string	true	"bearer token"
//	@Produce		json
//	@Success		200	{object}	Response[User]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/user [get]
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

type ChangeRoleByIDReq struct {
	RoleID int64 `json:"role_id"`
}

// ChangeRoleByID
//
//	@Summary		Change Role
//	@Description	admin change user role
//	@Tags			accounts
//	@Accept			json
//	@Param			Authorization	header	string				true	"bearer token"
//	@Param			id				path	int					true	"user id"
//	@Param			request			body	ChangeRoleByIDReq	true	"body request"
//	@Produce		json
//	@Success		200	{object}	Response[User]
//	@Failure		400	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/admin/user/{id} [put]
func (h *UserHandler) ChangeRoleByID(c echo.Context) error {
	ctx := c.Request().Context()

	ID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return NewAPIErr(c, NewErr(ErrInput, err, "id invalid"))
	}

	var input ChangeRoleByIDReq
	if err := c.Bind(&input); err != nil {
		return NewAPIErr(c, err)
	}
	c.Set(KeyInput, input)

	var user *User
	err = h.trxProvider.Transact(ctx, func(service *ServiceRegistry) error {
		user, err = service.User.ChangeRoleByID(ctx, int64(ID), UserInput{RoleID: input.RoleID})
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

// GetRoleByID
//
//	@Summary		Get Role
//	@Description	admin get role by id
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization			header		string	true	"bearer token"
//	@Param			id						path		int		true	"role id"
//	@Success		200						{object}	Response[Role]
//	@Failure		400						{object}	Response[any]
//	@Failure		500						{object}	Response[any]
//	@Router			/api/admin/roles/{id} 	[get]
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

// PaginationRole
//
//	@Summary		Filter Role
//	@Description	admin filter roles
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"bearer token"
//	@Param			page			query		int			false	"pagination page"
//	@Param			per_page		query		int			false	"pagination page size"
//	@Param			request			body		RoleFilter	false	"filter"
//	@Success		200				{object}	Response[Paginate[Role]]
//	@Failure		400				{object}	Response[any]
//	@Failure		500				{object}	Response[any]
//	@Router			/api/admin/roles/filter [post]
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
