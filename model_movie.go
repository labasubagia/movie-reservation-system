package main

import (
	"net/url"
	"strings"
	"time"
)

type MovieInput struct {
	Title       string    `json:"title,omitempty"`
	ReleaseDate time.Time `json:"release_date,omitempty" example:"2006-01-02T15:04:05+08:00"`
	Director    string    `json:"director,omitempty"`
	Duration    int64     `json:"duration,omitempty"`
	PosterURL   string    `json:"poster_url,omitempty"`
	Description string    `json:"description,omitempty"`
	GenreIDs    []int64   `json:"genre_ids,omitempty"`
}

func (i *MovieInput) Validate() error {
	i.Title = strings.Trim(i.Title, " ")
	i.Director = strings.Trim(i.Director, " ")
	i.PosterURL = strings.Trim(i.PosterURL, " ")
	i.Description = strings.Trim(i.Description, " ")

	if i.Title == "" {
		return NewErr(ErrInput, nil, "title is required")
	}
	if i.ReleaseDate.IsZero() {
		return NewErr(ErrInput, nil, "release date is required")
	}
	if i.Director == "" {
		return NewErr(ErrInput, nil, "director is required")
	}
	if i.Duration <= 0 {
		return NewErr(ErrInput, nil, "duration (minutes) is required")
	}
	if i.Description == "" {
		return NewErr(ErrInput, nil, "description is required")
	}
	if i.PosterURL == "" {
		return NewErr(ErrInput, nil, "poster url is required")
	}
	_, err := url.ParseRequestURI(i.PosterURL)
	if err != nil {
		return NewErr(ErrInput, err, "poster url is not a valid url")
	}
	if len(i.GenreIDs) == 0 {
		return NewErr(ErrInput, nil, "genre ids is required")
	}

	return nil
}

type MovieFilter struct {
	IDs           []int64   `json:"ids"`
	Search        string    `json:"search"`
	GenreIDs      []int64   `json:"genre_ids"`
	ShowtimeAfter time.Time `json:"showtime_after" example:"2006-01-02T15:04:05+08:00"`
}

func (f *MovieFilter) Validate() error {
	minSearch := 3
	if len(f.Search) > 0 && len(f.Search) < minSearch {
		return NewErr(ErrInput, nil, "minimum search character is %d", minSearch)
	}
	return nil
}

func NewMovie(input MovieInput) (*Movie, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	movie := Movie{
		Title:       input.Title,
		ReleaseDate: input.ReleaseDate,
		Director:    input.Director,
		Duration:    input.Duration,
		PosterURL:   input.PosterURL,
		Description: input.Description,
		GenreIDs:    input.GenreIDs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return &movie, nil
}

type Movie struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date"`
	Director    string    `json:"director"`
	Duration    int64     `json:"duration"`
	PosterURL   string    `json:"poster_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// relation
	GenreIDs []int64  `json:"genre_ids"`
	Genres   []string `json:"genres"`
}

func (m *Movie) ValidateDuration(startAt, endAt time.Time) error {
	duration := endAt.Sub(startAt).Minutes()
	if duration < float64(m.Duration) {
		return NewErr(ErrInput, nil, "showtime duration less than movie duration")
	}
	return nil
}

func (m *Movie) GetDuration() time.Duration {
	return time.Duration(m.Duration * int64(time.Minute))
}

type GenreFilter struct {
	IDs   []int64  `json:"ids"`
	Names []string `json:"names"`
}

func (i *GenreFilter) Validate() error {
	return nil
}

type GenreInput struct {
	Name string `json:"name"`
}

func (i *GenreInput) Validate() error {
	i.Name = strings.Trim(i.Name, " ")
	if i.Name == "" {
		return NewErr(ErrInput, nil, "name is required")
	}
	i.Name = strings.ToLower(i.Name)
	return nil
}

func NewGenre(input GenreInput) (*Genre, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}
	genre := Genre{Name: input.Name}
	return &genre, nil
}

type Genre struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
