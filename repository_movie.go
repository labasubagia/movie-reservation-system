package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewMovieRepository(tx pgx.Tx) *MovieRepository {
	return &MovieRepository{
		tx: tx,
	}
}

type MovieRepository struct {
	tx pgx.Tx
}

func (r *MovieRepository) Create(ctx context.Context, movie *Movie) (int64, error) {
	if len(movie.GenreIDs) == 0 {
		return 0, NewErr(ErrInput, nil, "genre ids is required")
	}

	sql := `
		insert into public.movies (title, release_date, director, duration, poster_url, description)
		values (@title, @release_date, @director, @duration, @poster_url, @description)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(
		ctx,
		sql,
		pgx.NamedArgs{
			"title":        movie.Title,
			"release_date": movie.ReleaseDate,
			"director":     movie.Director,
			"duration":     movie.Duration,
			"poster_url":   movie.PosterURL,
			"description":  movie.Description,
		},
	).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}

	sql = `
		insert into movie_genres (movie_id, genre_id)
		select @id, g.id from public.genres g where g.id = any(@genre_ids)
	`
	_, err = r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID, "genre_ids": movie.GenreIDs})
	if err != nil {
		return 0, NewSQLErr(err)
	}

	return ID, nil
}

func (r *MovieRepository) UpdateByID(ctx context.Context, ID int64, input MovieInput) error {
	sql := `
		update public.movies
		set
			title=@title,
			release_date=@release_date,
			director=@director,
			duration=@duration,
			poster_url=@poster_url,
			description=@description,
			updated_at=NOW()
		where id=@id
	`
	_, err := r.tx.Exec(
		ctx,
		sql,
		pgx.NamedArgs{
			"id":           ID,
			"title":        input.Title,
			"release_date": input.ReleaseDate,
			"director":     input.Director,
			"duration":     input.Duration,
			"poster_url":   input.PosterURL,
			"description":  input.Description,
		},
	)
	if err != nil {
		return NewSQLErr(err)
	}

	if len(input.GenreIDs) == 0 {
		return NewErr(ErrInput, nil, "genre ids is required")
	}

	sql = `delete from public.movie_genres mg where mg.movie_id=@id and mg.genre_id != any(@genre_ids)`
	_, err = r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID, "genre_ids": input.GenreIDs})
	if err != nil {
		return NewSQLErr(err)
	}

	// insert only if not exists yet
	sql = `
		insert into movie_genres (movie_id, genre_id)
		select @id, g.id
			from genres g
			left join movie_genres mg on g.id = mg.genre_id and mg.movie_id = @id
			where mg.genre_id is null and g.id = any(@genre_ids)
		on conflict do nothing
	`
	_, err = r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID, "genre_ids": input.GenreIDs})
	if err != nil {
		return NewSQLErr(err)
	}

	return nil
}

func (r *MovieRepository) DeleteByID(ctx context.Context, ID int64) error {
	sql := `delete from public.movies where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *MovieRepository) FindOne(ctx context.Context, filter MovieFilter) (*Movie, error) {
	movies, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(movies) == 0 {
		return nil, NewErr(ErrNotFound, nil, "movie not found")
	}
	return &movies[0], nil
}

func (r *MovieRepository) Find(ctx context.Context, filter MovieFilter) ([]Movie, error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)
	sql := fmt.Sprintf(`
		select
			m.id,
			m.title,
			m.release_date,
			m.director,
			m.duration,
			m.poster_url,
			m.description,
			m.created_at,
			m.updated_at
		from
			public.movies m
		where
			m.id in (%s)
		order by
			m.release_date desc,
			m.id desc
	`, filterSQL)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var movies []Movie
	var movieIDs []int64
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.Director,
			&movie.Duration,
			&movie.PosterURL,
			&movie.Description,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		movies = append(movies, movie)
		movieIDs = append(movieIDs, movie.ID)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	sql = `
		select
			mg.movie_id,
			mg.genre_id,
			g.name
		from
			movie_genres mg
		join genres g on
			g.id = mg.genre_id
		where
			mg.movie_id = any(@movie_ids)
	`
	rows, err = r.tx.Query(ctx, sql, pgx.NamedArgs{"movie_ids": movieIDs})
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	genreMap := map[int64]struct {
		IDs   []int64
		Names []string
	}{}
	for rows.Next() {
		var movieID int64
		var genre Genre
		err := rows.Scan(
			&movieID,
			&genre.ID,
			&genre.Name,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		cur := genreMap[movieID]
		cur.IDs = append(cur.IDs, genre.ID)
		cur.Names = append(cur.Names, genre.Name)
		genreMap[movieID] = cur
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	for i := 0; i < len(movies); i++ {
		cur := movies[i]
		movies[i].GenreIDs = genreMap[cur.ID].IDs
		movies[i].Genres = genreMap[cur.ID].Names
	}

	return movies, nil
}

func (r *MovieRepository) Paginate(ctx context.Context, filter MovieFilter, page PaginateInput) (*Paginate[Movie], error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf("select count(*) from (%s)", filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Movie{}, totalItems, page.Page, page.Size)
	if totalItems == 0 {
		return p, nil
	}

	if page.Page > p.TotalPage {
		page.Page = p.TotalPage
		p.CurrentPage = page.Page
	}

	sql = fmt.Sprintf(`
		select
			m.id,
			m.title,
			m.release_date,
			m.director,
			m.duration,
			m.poster_url,
			m.description,
			m.created_at,
			m.updated_at
		from
			public.movies m
		where
			m.id in (%s)
		order by
			m.release_date desc,
			m.id desc
		limit @page_size offset (@page - 1) * @page_size
	`, filterSQL)
	rows, err := r.tx.Query(ctx, sql, mergeNamedArgs(
		filterArgs,
		pgx.NamedArgs{
			"page":      page.Page,
			"page_size": page.Size,
		}),
	)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var movies []Movie
	var movieIDs []int64
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.Director,
			&movie.Duration,
			&movie.PosterURL,
			&movie.Description,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		movies = append(movies, movie)
		movieIDs = append(movieIDs, movie.ID)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	sql = `
		select
			mg.movie_id,
			mg.genre_id,
			g.name
		from
			movie_genres mg
		join genres g on
			g.id = mg.genre_id
		where
			mg.movie_id = any(@movie_ids)
	`
	rows, err = r.tx.Query(ctx, sql, pgx.NamedArgs{"movie_ids": movieIDs})
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	genreMap := map[int64]struct {
		IDs   []int64
		Names []string
	}{}
	for rows.Next() {
		var movieID int64
		var genre Genre
		err := rows.Scan(
			&movieID,
			&genre.ID,
			&genre.Name,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		cur := genreMap[movieID]
		cur.IDs = append(cur.IDs, genre.ID)
		cur.Names = append(cur.Names, genre.Name)
		genreMap[movieID] = cur
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	for i := 0; i < len(movies); i++ {
		cur := movies[i]
		movies[i].GenreIDs = genreMap[cur.ID].IDs
		movies[i].Genres = genreMap[cur.ID].Names
	}

	p.Items = movies
	return p, nil
}

func (r *MovieRepository) getFilterSQL(_ context.Context, filter MovieFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _m.id
		from public.movies _m
		left join public.movie_genres _mg on _m.id = _mg.movie_id
		left join public.showtimes _s on _m.id = _s.movie_id
		where
			case
				when plainto_tsquery('simple', @_search)::text != '' then
					_m.search_vector @@ plainto_tsquery('simple', @_search)
				else
					true
			end
			and
			case
				when array_length(@_ids::int[], 1) > 0 then
					_m.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_genre_ids::int[], 1) > 0 then
					_mg.genre_id = any(@_genre_ids)
				else
					true
			end
			and
			case
				when @_showtime_after::timestamptz is not null then
					@_showtime_after <= _s.start_at
				else
					true
			end

	`
	args = pgx.NamedArgs{
		"_search":    filter.Search,
		"_ids":       filter.IDs,
		"_genre_ids": filter.GenreIDs,
	}
	if !filter.ShowtimeAfter.IsZero() {
		args["_showtime_after"] = filter.ShowtimeAfter
	}

	return sql, args
}

func (r *MovieRepository) CreateGenre(ctx context.Context, genre *Genre) (int64, error) {
	sql := `insert into public.genres (name) values (@name) returning id`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{"name": genre.Name}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *MovieRepository) UpdateGenreByID(ctx context.Context, ID int64, input GenreInput) error {
	sql := `update public.genres set name=@name where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":   ID,
		"name": input.Name,
	})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *MovieRepository) DeleteGenreByID(ctx context.Context, ID int64) error {
	sql := `delete from public.genres where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *MovieRepository) FindOneGenre(ctx context.Context, filter GenreFilter) (*Genre, error) {
	genres, err := r.FindGenre(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(genres) == 0 {
		return nil, NewErr(ErrNotFound, nil, "genre not found")
	}
	return &genres[0], nil
}

func (r *MovieRepository) FindGenre(ctx context.Context, filter GenreFilter) ([]Genre, error) {
	filterSQL, filterArgs := r.getGenreFilterSQL(ctx, filter)
	sql := fmt.Sprintf(`
		select
			g.id,
			g.name,
			g.created_at,
			g.updated_at
		from
			public.genres g
		where
			g.id in (%s)
		order by
			g.name asc,
			g.id desc
	`, filterSQL)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var genre Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		genres = append(genres, genre)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return genres, nil
}

func (r *MovieRepository) PaginateGenres(ctx context.Context, filter GenreFilter, page PaginateInput) (*Paginate[Genre], error) {
	filterSQL, filterArgs := r.getGenreFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf("select count(*) from (%s)", filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Genre{}, totalItems, page.Page, page.Size)
	if totalItems == 0 {
		return p, nil
	}

	if page.Page > p.TotalPage {
		page.Page = p.TotalPage
		p.CurrentPage = page.Page
	}

	sql = fmt.Sprintf(`
		select
			g.id,
			g.name,
			g.created_at,
			g.updated_at
		from
			public.genres g
		where
			g.id in (%s)
		order by
			g.name asc,
			g.id desc
		limit @page_size offset (@page - 1) * @page_size
	`, filterSQL)
	rows, err := r.tx.Query(ctx, sql, mergeNamedArgs(
		filterArgs,
		pgx.NamedArgs{
			"page":      page.Page,
			"page_size": page.Size,
		}),
	)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var genre Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		genres = append(genres, genre)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	p.Items = genres
	return p, nil
}

func (r *MovieRepository) getGenreFilterSQL(_ context.Context, filter GenreFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _g.id
		from genres _g
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_g.id = any(@_ids)
				else
					true
			end
			and
				case
					when array_length(@_names::text[], 1) > 0 then
						_g.name = any(@_names)
					else
						true
				end
	`
	args = pgx.NamedArgs{
		"_ids":   filter.IDs,
		"_names": filter.Names,
	}
	return sql, args
}
