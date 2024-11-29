package main

import "time"

type ShowtimeFilter struct {
	IDs      []int64   `json:"ids"`
	MovieIDs []int64   `json:"movie_ids"`
	RoomIDs  []int64   `json:"room_ids"`
	After    time.Time `json:"after"` // only list showtime after/equal this time
}

func (f *ShowtimeFilter) Validate() error {
	return nil
}

type ShowtimeInput struct {
	MovieID int64     `json:"movie_id,omitempty"`
	RoomID  int64     `json:"room_id,omitempty"`
	StartAt time.Time `json:"start_at,omitempty"`
	EndAt   time.Time `json:"end_at,omitempty"`
	Price   int       `json:"price,omitempty"`
}

func (i *ShowtimeInput) Validate() error {
	if i.MovieID <= 0 {
		return NewErr(ErrInput, nil, "movie id is invalid")
	}
	if i.RoomID <= 0 {
		return NewErr(ErrInput, nil, "room id is invalid")
	}
	if i.StartAt.IsZero() {
		return NewErr(ErrInput, nil, "start at is required")
	}
	if i.EndAt.IsZero() {
		return NewErr(ErrInput, nil, "end at is required")
	}
	if i.StartAt.After(i.EndAt) {
		return NewErr(ErrInput, nil, "time invalid")
	}
	if i.Price <= 0 {
		return NewErr(ErrInput, nil, "price minimum is 0")
	}
	return nil
}

func NewShowtime(input ShowtimeInput) (*Showtime, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	showtime := Showtime{
		MovieID: input.MovieID,
		RoomID:  input.RoomID,
		StartAt: input.StartAt,
		EndAt:   input.EndAt,
		Price:   input.Price,
	}
	return &showtime, nil
}

type Showtime struct {
	ID        int64     `json:"id,omitempty"`
	MovieID   int64     `json:"movie_id,omitempty"`
	RoomID    int64     `json:"room_id,omitempty"`
	StartAt   time.Time `json:"start_at,omitempty"`
	EndAt     time.Time `json:"end_at,omitempty"`
	Price     int       `json:"price,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// relation
	MovieTitle    string `json:"movie_title"`
	RoomName      string `json:"room_name"`
	TotalSeat     int64  `json:"total_seat"`
	AvailableSeat int64  `json:"available_seat"`
}

func (s *Showtime) ValidateOtherOverlapping(startAt, endAt time.Time) error {
	errOverlapping := NewErr(ErrInput, nil, "showtime room overlapping with other showtime")

	if s.StartAt.Equal(startAt) {
		return errOverlapping
	}
	if s.StartAt.Equal(startAt) {
		return errOverlapping
	}
	if startAt.After(s.StartAt) && startAt.Before(s.EndAt) {
		return errOverlapping
	}

	if s.EndAt.Equal(startAt) {
		return errOverlapping
	}
	if s.EndAt.Equal(endAt) {
		return errOverlapping
	}
	if endAt.After(s.StartAt) && endAt.Before(s.EndAt) {
		return errOverlapping
	}

	return nil
}
