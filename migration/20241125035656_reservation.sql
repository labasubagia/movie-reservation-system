-- +goose Up
-- +goose StatementBegin
CREATE TYPE public.reservation_status AS enum ('unpaid', 'paid', 'cancelled');
CREATE TABLE IF NOT EXISTS public.reservations (
    id bigserial NOT NULL,
    user_id bigint NOT NULL,
    showtime_id bigint NOT NULL,
    seat_id bigint NOT NULL,
    created_at timestamptz DEFAULT NOW() NOT NULL,
    updated_at timestamptz DEFAULT NOW() NOT NULL,
    status public.reservation_status DEFAULT 'unpaid' NOT NULL,
    CONSTRAINT reservations_pk PRIMARY KEY (id),
    CONSTRAINT reservations_users_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT reservations_showtimes_fk FOREIGN KEY (showtime_id) REFERENCES public.showtimes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT reservations_seats_fk FOREIGN KEY (seat_id) REFERENCES public.seats(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.reservations;
DROP TYPE IF EXISTS public.reservation_status;
-- +goose StatementEnd
