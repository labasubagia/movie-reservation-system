package main

import (
	"strings"
	"time"
)

type RoomFilter struct {
	IDs      []int64  `json:"ids"`
	Names    []string `json:"names"`
	IsUsable *bool    `json:"is_usable"`
}

func (f *RoomFilter) Validate() error {
	for i, v := range f.Names {
		name := strings.Trim(v, " ")
		if name == "" {
			return NewErr(ErrInput, nil, "name is required")
		}
		f.Names[i] = name
	}
	return nil
}

type RoomInput struct {
	Name string `json:"name,omitempty"`
}

func (i *RoomInput) Validate() error {
	i.Name = strings.Trim(i.Name, " ")
	if i.Name == "" {
		return NewErr(ErrInput, nil, "name is required")
	}
	return nil
}

type Room struct {
	ID        int64     `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// relation
	Capacity int64 `json:"capacity"`
}

func NewRoom(input RoomInput) (*Room, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	room := Room{
		Name: input.Name,
	}
	return &room, nil
}

type SeatInput struct {
	RoomID          int64  `json:"room_id,omitempty"`
	Name            string `json:"name,omitempty"`
	AdditionalPrice int    `json:"additional_price,omitempty"`
}

func (i *SeatInput) Validate() error {
	i.Name = strings.Trim(i.Name, " ")

	if i.Name == "" {
		return NewErr(ErrInput, nil, "name is required")
	}
	if i.RoomID <= 0 {
		return NewErr(ErrInput, nil, "room id is invalid")
	}
	if i.AdditionalPrice < 0 {
		return NewErr(ErrInput, nil, "additional price minimum is 0")
	}
	return nil
}

type SeatFilter struct {
	IDs     []int64  `json:"ids,omitempty"`
	RoomIDs []int64  `json:"room_ids,omitempty"`
	Names   []string `json:"names,omitempty"`
}

type Seat struct {
	ID              int64     `json:"id"`
	RoomID          int64     `json:"room_id"`
	Name            string    `json:"name"`
	AdditionalPrice int       `json:"additional_price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// relation
	IsAvailable bool `json:"is_available"`
}

func NewSeat(input SeatInput) (*Seat, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	seat := Seat{
		RoomID:          input.RoomID,
		Name:            strings.Trim(input.Name, " "),
		AdditionalPrice: input.AdditionalPrice,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return &seat, nil
}
