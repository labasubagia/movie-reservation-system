package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerHost string
	ServerPort int

	JWTSecret string

	PostgresHost     string
	PostgresPort     int64
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}

func NewConfig() *Config {
	c := Config{
		ServerHost: "localhost",
		ServerPort: 8000,

		JWTSecret: "secret",

		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "root",
		PostgresPassword: "root",
		PostgresDB:       "movie_reservation_system",
	}

	if value, err := strconv.Atoi(os.Getenv("SERVER_PORT")); err == nil {
		c.ServerPort = value
	}

	if value := os.Getenv("JWT_SECRET"); value != "" {
		c.JWTSecret = value
	}

	if value := os.Getenv("POSTGRES_HOST"); value != "" {
		c.PostgresHost = value
	}
	if value, err := strconv.Atoi(os.Getenv("POSTGRES_PORT")); err == nil {
		c.PostgresPort = int64(value)
	}
	if value := os.Getenv("POSTGRES_USER"); value != "" {
		c.PostgresUser = value
	}
	if value := os.Getenv("POSTGRES_PASSWORD"); value != "" {
		c.PostgresPassword = value
	}
	if value := os.Getenv("POSTGRES_DB"); value != "" {
		c.PostgresDB = value
	}
	return &c
}
