package main

import (
	"time"
)

type CartFilter struct {
	IDs         []int64 `json:"ids,omitempty"`
	UserIDs     []int64 `json:"user_ids,omitempty"`
	ShowtimeIDs []int64 `json:"showtime_ids,omitempty"`
	SeatIDs     []int64 `json:"seat_ids,omitempty"`
}

func (f *CartFilter) Validate() error {
	return nil
}

type CartInput struct {
	UserID     int64 `json:"user_id,omitempty"`
	ShowtimeID int64 `json:"showtime_id,omitempty"`
	SeatID     int64 `json:"seat_id,omitempty"`
}

func (i *CartInput) Validate() error {
	if i.UserID <= 0 {
		return NewErr(ErrInput, nil, "user id is invalid")
	}
	if i.ShowtimeID <= 0 {
		return NewErr(ErrInput, nil, "showtime id is invalid")
	}
	if i.SeatID <= 0 {
		return NewErr(ErrInput, nil, "seat id is invalid")
	}
	return nil
}

func NewCart(input CartInput) (*Cart, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	Cart := Cart{
		UserID:     input.UserID,
		ShowtimeID: input.ShowtimeID,
		SeatID:     input.SeatID,
	}
	return &Cart, nil
}

type Cart struct {
	ID         int64     `json:"id,omitempty"`
	UserID     int64     `json:"user_id,omitempty"`
	ShowtimeID int64     `json:"showtime_id,omitempty"`
	SeatID     int64     `json:"seat_id,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`

	// relation
	Movie         string    `json:"movie"`
	ShowtimeStart time.Time `json:"showtime_start"`
	ShowtimeEnd   time.Time `json:"showtime_end"`
	Room          string    `json:"room"`
	Seat          string    `json:"seat"`
	Price         int64     `json:"price"`
}
