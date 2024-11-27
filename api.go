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
	"golang.org/x/time/rate"
)

func RunServer(config *Config, handler *HandlerRegistry) error {
	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		NewAPIErr(c, err)
	}

	e.Use(middleware.Secure())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(5))))

	Route(e, config, handler)

	e.GET("/", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(200, "ok")
	})

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
