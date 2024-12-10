# Movie Reservation System

![build](https://github.com/labasubagia/movie-reservation-system/actions/workflows/go.yml/badge.svg)


## Description

This is the Backend API Movie Reservation System. It is an application designed to streamline the management and booking of movie showtimes.

This project is based on [roadmap.sh](https://roadmap.sh/projects/movie-reservation-system).

Visit the [Project Wiki] (https://github.com/labasubagia/movie-reservation-system/wiki) for detailed information, such as the user manual, architecture, technical details, etc.

## Installation

### Run Project
Run
```sh
# run docker compose
docker compose up -d

# install dependencies
go mod tidy

# run
go run .
```

Test
```sh
go test -v ./...
```


### API Docs

Visit /swagger/index.html

default is  http://localhost:8000/swagger/index.html

## License
[MIT](./LICENSE)
