package main

type ServiceRegistry struct {
	User        *UserService
	Movie       *MovieService
	Room        *RoomService
	Showtime    *ShowtimeService
	Reservation *ReservationService
	Cart        *CartService
}

func NewService(config *Config, repo *RepositoryRegistry) *ServiceRegistry {
	service := ServiceRegistry{
		User:        NewUserService(config, repo),
		Movie:       NewMovieService(config, repo),
		Room:        NewRoomService(config, repo),
		Showtime:    NewShowtimeService(config, repo),
		Reservation: NewReservationService(config, repo),
		Cart:        NewCartService(config, repo),
	}
	return &service
}
