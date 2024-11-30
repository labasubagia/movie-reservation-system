-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.carts (
    id bigserial NOT NULL,
    user_id bigint NOT NULL,
    showtime_id bigint NOT NULL,
    seat_id bigint NOT NULL,
    created_at timestamptz DEFAULT NOW() NOT NULL,
    updated_at timestamptz DEFAULT NOW() NOT NULL,
    CONSTRAINT carts_pk PRIMARY KEY (id),
    CONSTRAINT carts_users_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT carts_showtimes_fk FOREIGN KEY (showtime_id) REFERENCES public.showtimes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT carts_seats_fk FOREIGN KEY (seat_id) REFERENCES public.seats(id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE UNIQUE INDEX carts_unique_idx ON public.carts (user_id, showtime_id, seat_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.carts;
-- +goose StatementEnd
