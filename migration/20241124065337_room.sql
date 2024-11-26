-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.rooms (
	id bigserial NOT NULL,
	"name" varchar NOT NULL,
	created_at timestamptz DEFAULT NOW() NOT NULL,
	updated_at timestamptz DEFAULT NOW() NOT NULL,
	CONSTRAINT rooms_pk PRIMARY KEY (id),
	CONSTRAINT rooms_unique UNIQUE ("name")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.rooms;
-- +goose StatementEnd
