package main

import (
	"strings"
	"time"
)

type ReservationStatus string

const (
	ReservationUnpaid    ReservationStatus = "unpaid"
	ReservationPaid      ReservationStatus = "paid"
	ReservationCancelled ReservationStatus = "cancelled"
)

func isReservationStatusValid(s ReservationStatus) bool {
	return s == ReservationUnpaid || s == ReservationPaid || s == ReservationCancelled
}

type ReservationFilter struct {
	IDs       []int64  `json:"ids,omitempty"`
	UserIDs   []int64  `json:"user_ids,omitempty"`
	Statuses  []string `json:"statuses,omitempty"`
	WithItems bool     `json:"with_items,omitempty"`
}

func (f *ReservationFilter) Validate() error {
	for i, v := range f.Statuses {
		status := ReservationStatus(v)
		if !isReservationStatusValid(status) {
			return NewErr(ErrInput, nil, "status with index %d invalid", i)
		}
	}
	return nil
}

type ReservationInput struct {
	UserID     int64             `json:"user_id,omitempty"`
	TotalPrice int64             `json:"total_price,omitempty"`
	Status     ReservationStatus `json:"status,omitempty"`
	CartIDs    []int64           `json:"cart_ids,omitempty"`
}

func (i *ReservationInput) Validate() error {
	if i.UserID <= 0 {
		return NewErr(ErrInput, nil, "user id is invalid")
	}
	for i, cartID := range i.CartIDs {
		if cartID <= 0 {
			return NewErr(ErrInput, nil, "cart id with index %d invalid: %v", i, cartID)
		}
	}

	i.Status = ReservationStatus(strings.Trim(string(i.Status), " "))
	if i.Status == "" {
		i.Status = ReservationUnpaid
	}
	if !isReservationStatusValid(i.Status) {
		return NewErr(ErrInput, nil, "status invalid")
	}
	return nil
}

func NewReservation(input ReservationInput) (*Reservation, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	reservation := Reservation{
		UserID:     input.UserID,
		TotalPrice: input.TotalPrice,
		Status:     input.Status,
	}
	return &reservation, nil
}

type Reservation struct {
	ID         int64             `json:"id,omitempty"`
	UserID     int64             `json:"user_id,omitempty"`
	Status     ReservationStatus `json:"status,omitempty"`
	TotalPrice int64             `json:"total_price,omitempty"`
	CreatedAt  time.Time         `json:"created_at,omitempty"`
	UpdatedAt  time.Time         `json:"updated_at,omitempty"`

	// relation
	Items []ReservationItem `json:"reservation_items,omitempty"`
}

type ReservationItemFilter struct {
	IDs            []int64 `json:"ids,omitempty"`
	UserIDs        []int64 `json:"user_ids,omitempty"`
	ReservationIDs []int64 `json:"reservation_ids,omitempty"`
	ShowtimeIDs    []int64 `json:"showtime_ids,omitempty"`
	SeatIDs        []int64 `json:"seat_ids,omitempty"`
}

func (f *ReservationItemFilter) Validate() error {
	return nil
}

type ReservationItemInput struct {
	ID            int64 `json:"id,omitempty"`
	ReservationID int64 `json:"reservation_id,omitempty"`
	UserID        int64 `json:"user_id,omitempty"`
	ShowtimeID    int64 `json:"showtime_id,omitempty"`
	TotalPrice    int64 `json:"total_price,omitempty"`
	SeatID        int64 `json:"seat_id,omitempty"`
}

func (i *ReservationItemInput) Validate() error {
	return nil
}

func NewReservationItem(input ReservationItemInput) (*ReservationItem, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	item := ReservationItem{
		ReservationID: input.ReservationID,
		UserID:        input.UserID,
		ShowtimeID:    input.ShowtimeID,
		SeatID:        input.SeatID,
		TotalPrice:    input.TotalPrice,
	}
	return &item, nil
}

type ReservationItem struct {
	ID            int64     `json:"id,omitempty"`
	ReservationID int64     `json:"reservation_id,omitempty"`
	UserID        int64     `json:"user_id,omitempty"`
	ShowtimeID    int64     `json:"showtime_id,omitempty"`
	SeatID        int64     `json:"seat_id,omitempty"`
	TotalPrice    int64     `json:"total_price,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`

	// relation
	Movie         string    `json:"movie"`
	ShowtimeStart time.Time `json:"showtime_start"`
	ShowtimeEnd   time.Time `json:"showtime_end"`
	Room          string    `json:"room"`
	Seat          string    `json:"seat"`
}
