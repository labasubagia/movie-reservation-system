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
	IDs         []int64  `json:"ids,omitempty"`
	UserIDs     []int64  `json:"user_ids,omitempty"`
	ShowtimeIDs []int64  `json:"showtime_ids,omitempty"`
	SeatIDs     []int64  `json:"seat_ids,omitempty"`
	Statuses    []string `json:"statuses,omitempty"`
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
	ShowtimeID int64             `json:"showtime_id,omitempty"`
	SeatID     int64             `json:"seat_id,omitempty"`
	Status     ReservationStatus `json:"status,omitempty"`
}

func (i *ReservationInput) Validate() error {
	if i.UserID <= 0 {
		return NewErr(ErrInput, nil, "user id is invalid")
	}
	if i.ShowtimeID <= 0 {
		return NewErr(ErrInput, nil, "showtime id is invalid")
	}
	if i.SeatID <= 0 {
		return NewErr(ErrInput, nil, "seat id is invalid")
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
		ShowtimeID: input.ShowtimeID,
		SeatID:     input.SeatID,
		Status:     input.Status,
	}
	return &reservation, nil
}

type Reservation struct {
	ID         int64             `json:"id,omitempty"`
	UserID     int64             `json:"user_id,omitempty"`
	ShowtimeID int64             `json:"showtime_id,omitempty"`
	SeatID     int64             `json:"seat_id,omitempty"`
	Status     ReservationStatus `json:"status,omitempty"`
	CreatedAt  time.Time         `json:"created_at,omitempty"`
	UpdatedAt  time.Time         `json:"updated_at,omitempty"`
}
