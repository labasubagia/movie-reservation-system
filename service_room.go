package main

import "context"

func NewRoomService(config *Config, repo *RepositoryRegistry) *RoomService {
	return &RoomService{
		config: config,
		repo:   repo,
	}
}

type RoomService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *RoomService) Create(ctx context.Context, input RoomInput) (*Room, error) {
	newRoom, err := NewRoom(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Room.Create(ctx, newRoom)
	if err != nil {
		return nil, err
	}

	room, err := s.repo.Room.FindOne(ctx, RoomFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) UpdateByID(ctx context.Context, ID int64, input RoomInput) (*Room, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	err = s.repo.Room.UpdateByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	room, err := s.repo.Room.FindOne(ctx, RoomFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) GetByID(ctx context.Context, ID int64) (*Room, error) {
	return s.repo.Room.FindOne(ctx, RoomFilter{IDs: []int64{ID}})
}

func (s *RoomService) DeleteByID(ctx context.Context, ID int64) error {
	err := s.repo.Room.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *RoomService) Pagination(ctx context.Context, filter RoomFilter, page PaginateInput) (*Paginate[Room], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Room.Pagination(ctx, filter, page)
}

func (s *RoomService) SetSeats(ctx context.Context, roomID int64, input []SeatInput) error {

	room, err := s.repo.Room.FindOne(ctx, RoomFilter{IDs: []int64{roomID}})
	if err != nil {
		return err
	}
	if room == nil {
		return NewErr(ErrNotFound, nil, "no room found for id %d", roomID)
	}

	seats := make([]Seat, 0, len(input))
	for _, item := range input {
		item.RoomID = roomID
		newSeat, err := NewSeat(item)
		if err != nil {
			return err
		}
		seats = append(seats, *newSeat)
	}

	err = s.repo.Room.SetSeats(ctx, roomID, seats)
	if err != nil {
		return err
	}

	return nil
}
