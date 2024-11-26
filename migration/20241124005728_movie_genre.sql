-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.movie_genres (
	id bigserial NOT NULL,
	movie_id bigint NOT NULL,
	genre_id bigint NOT NULL,
	created_at timestamptz DEFAULT NOW() NOT NULL,
	CONSTRAINT movie_genres_pk PRIMARY KEY (id),
	CONSTRAINT movie_genres_movies_fk FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE ON UPDATE CASCADE,
	CONSTRAINT movie_genres_genres_fk FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX movie_genres_unique_idx ON public.movie_genres USING btree (movie_id, genre_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.movie_genres;
-- +goose StatementEnd
