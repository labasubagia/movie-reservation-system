package main

import (
	"net/mail"
	"strings"
	"time"
)

// user

const (
	UserRegular = "user"
	UserAdmin   = "admin"
)

func isValidRole(role string) bool {
	return role == UserRegular || role == UserAdmin
}

type RoleFilter struct {
	IDs   []int64  `json:"role_ids,omitempty"`
	Names []string `json:"roles,omitempty"`
}

func (f *RoleFilter) Validate() error {
	for i, v := range f.Names {
		role := strings.Trim(v, " ")
		if !isValidRole(v) {
			return NewErr(ErrInput, nil, "role index %d = %s invalid", i, v)
		}
		f.Names[i] = role
	}
	return nil
}

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type UserFilter struct {
	IDs     []int64  `json:"ids,omitempty"`
	Emails  []string `json:"emails,omitempty"`
	RoleIDs []int64  `json:"role_ids,omitempty"`
	Roles   []string `json:"roles,omitempty"`
}

func (f *UserFilter) Validate() error {
	for i, v := range f.Roles {
		role := strings.Trim(v, " ")
		if !isValidRole(v) {
			return NewErr(ErrInput, nil, "role index %d = %s invalid", i, v)
		}
		f.Roles[i] = role
	}
	return nil
}

type UserInput struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	RoleID   int64  `json:"role_id,omitempty"`
}

func (i *UserInput) Validate() error {

	i.Email = strings.Trim(i.Email, " ")
	i.Password = strings.Trim(i.Password, " ")

	_, err := mail.ParseAddress(i.Email)
	if err != nil {
		return NewErr(ErrInput, nil, "email invalid!")
	}

	min := 8
	if len(i.Password) < min {
		return NewErr(ErrInput, nil, "password minimum %d characters", min)
	}

	max := 20
	if len(i.Password) > max {
		return NewErr(ErrInput, nil, "password maximum %d characters", max)
	}

	return nil
}

func NewUser(input UserInput) (*User, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user := User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return &user, nil
}

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Role string `json:"role,omitempty"`
}

func (u *User) UpdatePassword(password string) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) error {
	return VerifyPassword(password, u.PasswordHash)
}
