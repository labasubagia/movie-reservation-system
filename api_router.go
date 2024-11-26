package main

import "github.com/labstack/echo/v4"

func Route(e *echo.Echo, config *Config, handler *HandlerRegistry) {
	public := e.Group("/api")
	{
		public.POST("/register", handler.User.Register)
		public.POST("/login", handler.User.Login)
		public.GET("/movies", handler.Movie.Pagination)
	}

	loggedIn := e.Group("/api", jwtMiddleware(config))
	{
		loggedIn.GET("/user", handler.User.LoggedIn)

		loggedIn.GET("/showtimes/:id", handler.Showtime.GetByID)
		loggedIn.GET("/showtimes/:id/seats", handler.Showtime.GetShowtimeSeatByID)
		loggedIn.GET("/showtimes", handler.Showtime.Pagination)

		loggedIn.GET("/reservations/:id", handler.Reservation.UserGetByID)
		loggedIn.GET("/reservations", handler.Reservation.UserGetPagination)
		loggedIn.POST("/reservations", handler.Reservation.UserCreate)
		loggedIn.PUT("/reservations/:id", handler.Reservation.UserUpdateByID)
		loggedIn.DELETE("/reservations/:id", handler.Reservation.UserDeleteByID)
	}

	admin := e.Group("/api/admin", jwtMiddleware(config), adminMiddleware)
	{
		admin.PUT("/user/:id", handler.User.ChangeRoleByID)

		admin.GET("/genres", handler.Movie.PaginationGenre)
		admin.POST("/genres", handler.Movie.CreateGenre)
		admin.PUT("/genres/:id", handler.Movie.UpdateGenreByID)
		admin.DELETE("/genres/:id", handler.Movie.DeleteGenreByID)

		admin.GET("/movies", handler.Movie.Pagination)
		admin.GET("/movies/:id", handler.Movie.GetByID)
		admin.POST("/movies", handler.Movie.Create)
		admin.PUT("/movies/:id", handler.Movie.UpdateByID)
		admin.DELETE("/movies/:id", handler.Movie.DeleteByID)

		admin.GET("/rooms/:id", handler.Room.GetByID)
		admin.GET("/rooms", handler.Room.Pagination)
		admin.POST("/rooms", handler.Room.Create)
		admin.PUT("/rooms/:id", handler.Room.UpdateByID)
		admin.DELETE("/rooms/:id", handler.Movie.DeleteByID)

		admin.POST("/rooms/:id/seats", handler.Room.SetSeats)

		admin.GET("/showtimes/:id", handler.Showtime.GetByID)
		admin.GET("/showtimes", handler.Showtime.Pagination)
		admin.POST("/showtimes", handler.Showtime.Create)
		admin.PUT("/showtimes/:id", handler.Showtime.UpdateByID)
		admin.DELETE("/showtimes/:id", handler.Showtime.DeleteByID)
	}
}
