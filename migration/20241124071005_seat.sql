-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.seats (
    id bigserial NOT NULL,
    room_id bigint NOT NULL,
    "name" varchar NOT NULL,
    additional_price int DEFAULT 0 NOT NULL,
    created_at timestamptz DEFAULT NOW() NOT NULL,
    updated_at timestamptz DEFAULT NOW() NOT NULL,
    CONSTRAINT seats_pk PRIMARY KEY (id),
    CONSTRAINT seats_rooms_fk FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX seats_unique_idx ON public.seats (room_id,"name");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.seats;
-- +goose StatementEnd
