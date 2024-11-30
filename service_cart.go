package main

import "context"

func NewCartService(config *Config, repo *RepositoryRegistry) *CartService {
	return &CartService{
		config: config,
		repo:   repo,
	}
}

type CartService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *CartService) Create(ctx context.Context, input CartInput) (*Cart, error) {

	newCart, err := NewCart(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Cart.Create(ctx, newCart)
	if err != nil {
		return nil, err
	}

	cart, err := s.repo.Cart.FindOne(ctx, CartFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) UserUpdateByID(ctx context.Context, userID, ID int64, input CartInput) (*Cart, error) {
	input.UserID = userID
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	err = s.repo.Cart.UpdateByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	cart, err := s.repo.Cart.FindOne(ctx, CartFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) UserDeleteByID(ctx context.Context, userID, ID int64) error {
	_, err := s.repo.Cart.FindOne(ctx, CartFilter{IDs: []int64{ID}, UserIDs: []int64{userID}})
	if err != nil {
		return err
	}

	err = s.repo.Cart.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *CartService) UserGetByID(ctx context.Context, userID, ID int64) (*Cart, error) {
	return s.repo.Cart.FindOne(ctx, CartFilter{IDs: []int64{ID}, UserIDs: []int64{userID}})
}

func (s *CartService) Pagination(ctx context.Context, filter CartFilter, page PaginateInput) (*Paginate[Cart], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Cart.Pagination(ctx, filter, page)
}
