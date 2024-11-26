package main

import "context"

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

	olds, err := s.repo.Reservation.Find(ctx, ReservationFilter{
		UserIDs:     []int64{input.UserID},
		ShowtimeIDs: []int64{input.ShowtimeID},
		SeatIDs:     []int64{input.SeatID},
		Statuses:    []string{string(ReservationPaid), string(ReservationUnpaid)},
	})
	if err != nil {
		return nil, err
	}

	oldMap := map[ReservationStatus]struct{}{}
	for _, cur := range olds {
		oldMap[ReservationStatus(cur.Status)] = struct{}{}
	}
	if _, ok := oldMap[ReservationPaid]; ok {
		return nil, NewErr(ErrInput, nil, "you already paid this seat")
	}
	if _, ok := oldMap[ReservationUnpaid]; ok {
		return nil, NewErr(ErrInput, nil, "you already reserve this seat, please complete the transaction")
	}

	newReservation, err := NewReservation(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Reservation.Create(ctx, newReservation)
	if err != nil {
		return nil, err
	}

	reservation, err := s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *ReservationService) UserUpdateByID(ctx context.Context, userID, ID int64, input ReservationInput) (*Reservation, error) {
	input.UserID = userID
	err := input.Validate()
	if err != nil {
		return nil, err
	}

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

	err = s.repo.Reservation.UpdateByID(ctx, ID, input)
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
	return s.repo.Reservation.FindOne(ctx, ReservationFilter{IDs: []int64{ID}, UserIDs: []int64{userID}})
}

func (s *ReservationService) Pagination(ctx context.Context, filter ReservationFilter, page PaginateInput) (*Paginate[Reservation], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Reservation.Pagination(ctx, filter, page)
}
