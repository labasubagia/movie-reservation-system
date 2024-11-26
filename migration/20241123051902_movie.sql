-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.movies (
	id bigserial NOT NULL,
	title varchar NOT NULL,
	release_date timestamptz NOT NULL,
	director varchar NOT NULL,
	duration int4 NOT NULL,
	poster_url varchar NOT NULL,
	description text NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	updated_at timestamptz DEFAULT now() NOT NULL,
	search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple'::regconfig, (title::text || ' '::text) || director::text)) STORED NULL,
	CONSTRAINT movies_pk PRIMARY KEY (id)
);
CREATE INDEX movies_search_vector_idx ON public.movies USING gin (search_vector);
CREATE UNIQUE INDEX movies_unique_idx ON public.movies USING btree (title, director);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.movies;
-- +goose StatementEnd
