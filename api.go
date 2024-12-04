package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"

	_ "movie-reservation-system/docs"
)

func RunServer(config *Config, handler *HandlerRegistry) error {
	e := setupServer(config, handler)
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(5))))
	e.Use(middleware.Secure())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		err := e.Start(fmt.Sprintf(":%d", config.ServerPort))
		if err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	return nil
}

//	@title			Movie Reservation System
//	@version		1.0
//	@description	This is API for movie reservation system.
//	@host			http://localhost:8000
//	@BasePath		/api
func setupServer(config *Config, handler *HandlerRegistry) *echo.Echo {
	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		NewAPIErr(c, err)
	}
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	Route(e, config, handler)
	return e
}
