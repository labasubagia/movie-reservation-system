package main

const (
	KeyInput  = "input"
	KeyOutput = "output"
)

type Response[T any] struct {
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type HandlerRegistry struct {
	User        *UserHandler
	Movie       *MovieHandler
	Room        *RoomHandler
	Showtime    *ShowtimeHandler
	Reservation *ReservationHandler
	Cart        *CartHandler
}

func NewHandler(config *Config, trxProvider *TransactionProvider) *HandlerRegistry {
	return &HandlerRegistry{
		User:        NewUserHandler(config, trxProvider),
		Movie:       NewMovieHandler(config, trxProvider),
		Room:        NewRoomHandler(config, trxProvider),
		Showtime:    NewShowtimeHandler(config, trxProvider),
		Reservation: NewReservationHandler(config, trxProvider),
		Cart:        NewCartHandler(config, trxProvider),
	}
}
