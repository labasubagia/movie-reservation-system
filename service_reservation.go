package main

import (
	"context"
	"time"
)

func NewReservationService(config *Config, repo *RepositoryRegistry) *ReservationService {
	return &ReservationService{
		config: config,
		repo:   repo,
	}
}

type ReservationService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *ReservationService) Create(ctx context.Context, input ReservationInput) (*Reservation, error) {

	err := input.Validate()
	if err != nil {
		return nil, err
	}

	carts, err := s.repo.Cart.Find(ctx, CartFilter{IDs: input.CartIDs, UserIDs: []int64{input.UserID}})
	if err != nil {
		return nil, err
	}
	if len(carts) == 0 {
		return nil, NewErr(ErrInput, nil, "no cart exists based on input")
	}

	var totalPrice int64
	showtimeSet := map[int64]struct{}{}
	for _, cart := range carts {
		showtimeSet[cart.ShowtimeID] = struct{}{}
		if len(showtimeSet) > 1 {
			return nil, NewErr(ErrInput, nil, "reservation can only created on cart with same showtime")
		}
		totalPrice += cart.Price
	}

	newReservation, err := NewReservation(input)
	if err != nil {
		return nil, err
	}
	newReservation.TotalPrice = totalPrice

	ID, err := s.repo.Reservation.Create(ctx, newReservation)
	if err != nil {
		return nil, err
	}

	reservation, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	for _, cart := range carts {
		item, err := NewReservationItem(ReservationItemInput{
			ReservationID: reservation.ID,
			UserID:        input.UserID,
			ShowtimeID:    cart.ShowtimeID,
			SeatID:        cart.SeatID,
			TotalPrice:    cart.Price,
		})
		if err != nil {
			return nil, err
		}
		_, err = s.repo.Reservation.CreateItem(ctx, item)
		if err != nil {
			return nil, err
		}

		err = s.repo.Cart.DeleteByID(ctx, cart.ID)
		if err != nil {
			return nil, err
		}
	}

	return reservation, nil
}

func (s *ReservationService) Pay(ctx context.Context, userID, ID int64) (*Reservation, error) {

	old, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{
		IDs:      []int64{ID},
		UserIDs:  []int64{userID},
		Statuses: []string{string(ReservationUnpaid)},
	})
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, NewErr(ErrInput, nil, "reservation not found")
	}

	err = s.repo.Reservation.UpdateByID(ctx, ID, ReservationInput{
		UserID:     userID,
		Status:     ReservationPaid,
		TotalPrice: old.TotalPrice,
	})
	if err != nil {
		return nil, err
	}

	reservation, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *ReservationService) Cancel(ctx context.Context, userID, ID int64) (*Reservation, error) {

	old, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{
		IDs:       []int64{ID},
		UserIDs:   []int64{userID},
		Statuses:  []string{string(ReservationUnpaid)},
		WithItems: true,
	})
	if err != nil {
		return nil, err
	}

	var minTime time.Time
	for _, item := range old.Items {
		if minTime.IsZero() {
			minTime = item.ShowtimeStart
		}
		if minTime.After(item.ShowtimeStart) {
			minTime = item.ShowtimeStart
		}
	}

	diff := minTime.Sub(time.Now())
	minimumTimeCancel := 6 * time.Hour
	if diff < minimumTimeCancel {
		return nil, NewErr(ErrInput, nil, "cannot cancel reservation below 6 hour showtime")
	}

	err = s.repo.Reservation.UpdateByID(ctx, ID, ReservationInput{
		UserID:     userID,
		Status:     ReservationCancelled,
		TotalPrice: old.TotalPrice,
	})
	if err != nil {
		return nil, err
	}

	reservation, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *ReservationService) UserDeleteByID(ctx context.Context, userID, ID int64) error {
	_, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}, UserIDs: []int64{userID}})
	if err != nil {
		return err
	}

	err = s.repo.Reservation.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ReservationService) UserGetByID(ctx context.Context, userID, ID int64) (*Reservation, error) {
	return s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}, UserIDs: []int64{userID}, WithItems: true})
}

func (s *ReservationService) Pagination(ctx context.Context, filter ReservationFilter, page PaginateInput) (*Paginate[Reservation], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Reservation.Pagination(ctx, filter, page)
}
