package main

import "github.com/jackc/pgx/v5"

type RepositoryRegistry struct {
	User        *UserRepository
	Movie       *MovieRepository
	Room        *RoomRepository
	Showtime    *ShowtimeRepository
	Reservation *ReservationRepository
}

func NewRepositoryRegistry(tx pgx.Tx) *RepositoryRegistry {
	return &RepositoryRegistry{
		User:        NewUserRepository(tx),
		Movie:       NewMovieRepository(tx),
		Room:        NewRoomRepository(tx),
		Showtime:    NewShowtimeRepository(tx),
		Reservation: NewReservationRepository(tx),
	}
}
