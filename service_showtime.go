package main

import "context"

func NewShowtimeService(config *Config, repo *RepositoryRegistry) *ShowtimeService {
	return &ShowtimeService{
		config: config,
		repo:   repo,
	}
}

type ShowtimeService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *ShowtimeService) Create(ctx context.Context, input ShowtimeInput) (*Showtime, error) {
	newShowtime, err := NewShowtime(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Showtime.Create(ctx, newShowtime)
	if err != nil {
		return nil, err
	}

	room, err := s.repo.Showtime.FindOne(ctx, ShowtimeFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *ShowtimeService) UpdateByID(ctx context.Context, ID int64, input ShowtimeInput) (*Showtime, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	err = s.repo.Showtime.UpdateByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	room, err := s.repo.Showtime.FindOne(ctx, ShowtimeFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *ShowtimeService) GetByID(ctx context.Context, ID int64) (*Showtime, error) {
	return s.repo.Showtime.FindOne(ctx, ShowtimeFilter{IDs: []int64{ID}})
}

func (s *ShowtimeService) DeleteByID(ctx context.Context, ID int64) error {
	err := s.repo.Showtime.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ShowtimeService) Pagination(ctx context.Context, filter ShowtimeFilter, page PaginateInput) (*Paginate[Showtime], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Showtime.Pagination(ctx, filter, page)
}

func (s *ShowtimeService) GetShowtimeSeats(ctx context.Context, showtimeID int64) ([]Seat, error) {
	return s.repo.Showtime.GetShowtimeSeats(ctx, showtimeID)
}
