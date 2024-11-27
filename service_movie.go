package main

import "context"

func NewMovieService(config *Config, repo *RepositoryRegistry) *MovieService {
	return &MovieService{
		config: config,
		repo:   repo,
	}
}

type MovieService struct {
	config *Config
	repo   *RepositoryRegistry
}

func (s *MovieService) Create(ctx context.Context, input MovieInput) (*Movie, error) {
	newMovie, err := NewMovie(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Movie.Create(ctx, newMovie)
	if err != nil {
		return nil, err
	}

	movie, err := s.repo.Movie.FindOne(ctx, MovieFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *MovieService) UpdateByID(ctx context.Context, ID int64, input MovieInput) (*Movie, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	err = s.repo.Movie.UpdateByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	movie, err := s.repo.Movie.FindOne(ctx, MovieFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *MovieService) DeleteByID(ctx context.Context, ID int64) error {
	err := s.repo.Movie.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *MovieService) GetByID(ctx context.Context, ID int64) (*Movie, error) {
	return s.repo.Movie.FindOne(ctx, MovieFilter{IDs: []int64{ID}})
}

func (s *MovieService) Pagination(ctx context.Context, filter MovieFilter, page PaginateInput) (*Paginate[Movie], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Movie.Paginate(ctx, filter, page)
}

func (s *MovieService) CreateGenre(ctx context.Context, input GenreInput) (*Genre, error) {
	newGenre, err := NewGenre(input)
	if err != nil {
		return nil, err
	}

	ID, err := s.repo.Movie.CreateGenre(ctx, newGenre)
	if err != nil {
		return nil, err
	}

	genre, err := s.repo.Movie.FindOneGenre(ctx, GenreFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return genre, nil
}

func (s *MovieService) UpdateGenreByID(ctx context.Context, ID int64, input GenreInput) (*Genre, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	err = s.repo.Movie.UpdateGenreByID(ctx, ID, input)
	if err != nil {
		return nil, err
	}

	genre, err := s.repo.Movie.FindOneGenre(ctx, GenreFilter{IDs: []int64{ID}})
	if err != nil {
		return nil, err
	}

	return genre, nil
}

func (s *MovieService) DeleteGenreByID(ctx context.Context, ID int64) error {
	err := s.repo.Movie.DeleteGenreByID(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *MovieService) GetGenreByID(ctx context.Context, ID int64) (*Genre, error) {
	return s.repo.Movie.FindOneGenre(ctx, GenreFilter{IDs: []int64{ID}})
}

func (s *MovieService) PaginationGenre(ctx context.Context, filter GenreFilter, page PaginateInput) (*Paginate[Genre], error) {
	err := filter.Validate()
	if err != nil {
		return nil, err
	}
	return s.repo.Movie.PaginateGenres(ctx, filter, page)
}
