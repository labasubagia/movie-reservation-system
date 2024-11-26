-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.genres (
	id serial NOT NULL,
	"name" varchar NOT NULL,
	created_at timestamptz DEFAULT NOW() NOT NULL,
	updated_at timestamptz DEFAULT NOW() NOT NULL,
	CONSTRAINT genres_pk PRIMARY KEY (id),
	CONSTRAINT genres_unique UNIQUE ("name")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.genres;
-- +goose StatementEnd
