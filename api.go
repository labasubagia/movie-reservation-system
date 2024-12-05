package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"

	"github.com/labasubagia/movie-reservation-system/docs"
	_ "github.com/labasubagia/movie-reservation-system/docs"
)

// RunServer
//
//	@title			Movie Reservation System
//	@version		1.0
//	@description	This is API for movie reservation system.
//	@BasePath		/api
func RunServer(config *Config, handler *HandlerRegistry) error {
	docs.SwaggerInfo.Host = config.ServerAddr()

	e := setupServer(config, handler)
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(5))))
	e.Use(middleware.Secure())
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		err := e.Start(config.ServerAddr())
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

func setupServer(config *Config, handler *HandlerRegistry) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		NewAPIErr(c, err)
	}
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	Route(e, config, handler)
	return e
}
