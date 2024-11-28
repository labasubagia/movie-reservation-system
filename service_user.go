package main

import (
	"context"
)

func NewUserService(config *Config, repo *RepositoryRegistry) *UserService {
	return &UserService{
		config: config,
		repo:   repo,
	}
}

type UserService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *UserService) Register(ctx context.Context, input UserInput) (*User, error) {
	existing, err := s.repo.User.FindOne(ctx, UserFilter{Emails: []string{input.Email}})
	if err != nil {
		if !ErrIs(err, ErrNotFound) {
			return nil, err
		}
	}
	if existing != nil {
		return nil, NewErr(ErrInput, nil, "user already exists!")
	}

	newUser, err := NewUser(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.User.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.User.FindOne(ctx, UserFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (s *UserService) Login(ctx context.Context, input UserInput) (string, error) {
	e := NewErr(ErrInput, nil, "email or password invalid!")

	user, err := s.repo.User.FindOne(ctx, UserFilter{Emails: []string{input.Email}})
	if err != nil {
		if ErrIs(err, ErrNotFound) {
			return "", e
		}
		return "", err
	}

	err = user.VerifyPassword(input.Password)
	if err != nil {
		return "", e
	}

	token, err := CreateToken(s.config, user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.User.FindOne(ctx, UserFilter{Emails: []string{email}})
}

func (s *UserService) ChangeRoleByID(ctx context.Context, ID int64, input UserInput) (*User, error) {
	err := s.repo.User.UpdateRoleByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.User.FindOne(ctx, UserFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetRoleByID(ctx context.Context, ID int64) (*Role, error) {
	return s.repo.User.FindRoleOne(ctx, RoleFilter{IDs: []int64{ID}})
}

func (s *UserService) PaginationRole(ctx context.Context, filter RoleFilter, page PaginateInput) (*Paginate[Role], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.User.PaginateRole(ctx, filter, page)
}
