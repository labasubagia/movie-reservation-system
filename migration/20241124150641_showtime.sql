-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.showtimes (
    id bigserial NOT NULL,
    movie_id bigint NOT NULL,
    room_id bigint NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    price int NOT NULL,
    created_at timestamptz DEFAULT NOW() NOT NULL,
    updated_at timestamptz DEFAULT NOW() NOT NULL,
    CONSTRAINT showtimes_pk PRIMARY KEY (id),
    CONSTRAINT showtimes_movies_fk FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT showtimes_rooms_fk FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.showtimes;
-- +goose StatementEnd
