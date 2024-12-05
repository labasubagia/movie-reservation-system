package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewShowtimeRepository(tx pgx.Tx) *ShowtimeRepository {
	return &ShowtimeRepository{
		tx: tx,
	}
}

type ShowtimeRepository struct {
	tx pgx.Tx
}

func (r *ShowtimeRepository) Create(ctx context.Context, showtime *Showtime) (int64, error) {
	sql := `
		insert into public.showtimes (movie_id, room_id, start_at, end_at, price)
		values (@movie_id, @room_id, @start_at, @end_at, @price)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(
		ctx,
		sql,
		pgx.NamedArgs{
			"movie_id": showtime.MovieID,
			"room_id":  showtime.RoomID,
			"start_at": showtime.StartAt,
			"end_at":   showtime.EndAt,
			"price":    showtime.Price,
		},
	).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *ShowtimeRepository) UpdateByID(ctx context.Context, ID int64, input ShowtimeInput) error {
	sql := `
		update
			public.showtimes
		set
			updated_at = now(),
			movie_id = @movie_id,
			room_id = @room_id,
			start_at = @start_at,
			end_at = @end_at,
			price = @price
		where
			id = @id
	`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":       ID,
		"movie_id": input.MovieID,
		"room_id":  input.RoomID,
		"start_at": input.StartAt,
		"end_at":   input.EndAt,
		"price":    input.Price,
	})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *ShowtimeRepository) DeleteByID(ctx context.Context, ID int64) error {
	sql := `delete from showtimes where id = @id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *ShowtimeRepository) FindOne(ctx context.Context, filter ShowtimeFilter) (*Showtime, error) {
	showtimes, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(showtimes) == 0 {
		return nil, NewErr(ErrNotFound, nil, "showtime not found")
	}
	return &showtimes[0], nil
}

func (r *ShowtimeRepository) Find(ctx context.Context, filter ShowtimeFilter) ([]Showtime, error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	// TODO: fix this query
	sql := fmt.Sprintf(
		`
			select
				s.id,
				s.movie_id,
				s.room_id,
				s.start_at,
				s.end_at,
				s.price,
				s.created_at,
				s.updated_at,
				m.title as movie_title,
				r.name as room_name,
				count(st.*) as total_seat,
				count(st.*)-count(rv.*) as available_seat
			from
				showtimes s
			join movies m on
				m.id = s.movie_id
			join rooms r on
				r.id = s.room_id
			left join seats st on st.room_id  = r.id
			left join reservation_items rvi on rvi.showtime_id = s.id and rvi.seat_id = st.id
			left join reservations rv on rv.id = rvi.reservation_id and rv.status != 'cancelled'
			where s.id in (%s)
			group by s.id, m.title , r."name"
			order BY s.start_at asc, m.title asc
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var showtimes []Showtime
	for rows.Next() {
		var showtime Showtime
		err := rows.Scan(
			&showtime.ID,
			&showtime.MovieID,
			&showtime.RoomID,
			&showtime.StartAt,
			&showtime.EndAt,
			&showtime.Price,
			&showtime.CreatedAt,
			&showtime.UpdatedAt,
			&showtime.MovieTitle,
			&showtime.RoomName,
			&showtime.TotalSeat,
			&showtime.AvailableSeat,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		showtimes = append(showtimes, showtime)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return showtimes, nil
}

func (r *ShowtimeRepository) Pagination(ctx context.Context, filter ShowtimeFilter, page PaginateInput) (*Paginate[Showtime], error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf(`select count(*) from (%s)`, filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Showtime{}, totalItems, page.Page, page.Size)
	if totalItems == 0 {
		return p, nil
	}

	if page.Page > p.TotalPage {
		page.Page = p.TotalPage
		p.CurrentPage = page.Page
	}

	// TODO: fix this query
	sql = fmt.Sprintf(
		`
			select
				s.id,
				s.movie_id,
				s.room_id,
				s.start_at,
				s.end_at,
				s.price,
				s.created_at,
				s.updated_at,
				m.title as movie_title,
				r.name as room_name,
				count(st.*) as total_seat,
				count(st.*)-count(rv.*) as available_seat
			from
				showtimes s
			join movies m on
				m.id = s.movie_id
			join rooms r on
				r.id = s.room_id
			left join seats st on st.room_id  = r.id
			left join reservation_items rvi on rvi.showtime_id = s.id and rvi.seat_id = st.id
			left join reservations rv on rv.id = rvi.reservation_id and rv.status != 'cancelled'
			where s.id in (%s)
			group by s.id, m.title , r."name"
			order BY s.start_at asc, m.title asc
			limit @page_size offset (@page - 1) * @page_size;
		`,
		filterSQL,
	)
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

	var showtimes []Showtime
	for rows.Next() {
		var showtime Showtime
		err := rows.Scan(
			&showtime.ID,
			&showtime.MovieID,
			&showtime.RoomID,
			&showtime.StartAt,
			&showtime.EndAt,
			&showtime.Price,
			&showtime.CreatedAt,
			&showtime.UpdatedAt,
			&showtime.MovieTitle,
			&showtime.RoomName,
			&showtime.TotalSeat,
			&showtime.AvailableSeat,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		showtimes = append(showtimes, showtime)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p.Items = showtimes
	return p, nil
}

func (r *ShowtimeRepository) getFilterSQL(_ context.Context, filter ShowtimeFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _s.id
		from public.showtimes _s
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_s.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_movie_ids::int[], 1) > 0 then
					_s.movie_id = any(@_movie_ids)
				else
					true
			end
			and
			case
				when array_length(@_room_ids::int[], 1) > 0 then
					_s.room_id = any(@_room_ids)
				else
					true
			end
			and
			case
				when @_after::timestamptz is not null then
					@_after <= _s.start_at
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":       filter.IDs,
		"_movie_ids": filter.MovieIDs,
		"_room_ids":  filter.RoomIDs,
	}
	if !filter.After.IsZero() {
		args["_after"] = filter.After
	}
	return sql, args
}

// showtime seats

func (r *ShowtimeRepository) GetShowtimeSeats(ctx context.Context, showtimeID int64) ([]Seat, error) {
	sql := `
		with reserved_seat as (
		select
			ri.*
		from
			reservation_items ri
		join reservations r on
			ri.reservation_id = r.id
		where
			r.status = 'paid'::reservation_status
			or (r.status = 'unpaid'::reservation_status
				and r.created_at > now() - interval '30 minutes')
		)
		select
			st.id as seat_id,
			s.room_id,
			st.additional_price,
			st."name",
			st.created_at,
			st.updated_at,
			rt.id is null as is_available
		from
			showtimes s
		join seats st on
			s.room_id = st.room_id
		left join reserved_seat rt on
			rt.seat_id = st.id
			and rt.showtime_id = s.id
		where s.id = @showtime_id
		order by
			s.id,
			st.room_id,
			st."name"
	`
	rows, err := r.tx.Query(ctx, sql, pgx.NamedArgs{"showtime_id": showtimeID})
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var seat Seat
		err := rows.Scan(
			&seat.ID,
			&seat.RoomID,
			&seat.AdditionalPrice,
			&seat.Name,
			&seat.CreatedAt,
			&seat.UpdatedAt,
			&seat.IsAvailable,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		seats = append(seats, seat)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return seats, nil
}
