package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewCartRepository(tx pgx.Tx) *CartRepository {
	return &CartRepository{
		tx: tx,
	}
}

type CartRepository struct {
	tx pgx.Tx
}

func (r *CartRepository) Create(ctx context.Context, cart *Cart) (int64, error) {
	sql := `
		insert into public.carts (user_id, showtime_id, seat_id)
		values (@user_id, @showtime_id, @seat_id)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{
		"user_id":     cart.UserID,
		"showtime_id": cart.ShowtimeID,
		"seat_id":     cart.SeatID,
	}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *CartRepository) UpdateByID(ctx context.Context, ID int64, input CartInput) error {
	sql := `
		update public.carts
		set updated_at=now(), user_id=@user_id, showtime_id=@showtime_id, seat_id=@seat_id
		where id=@id
	`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":          ID,
		"user_id":     input.UserID,
		"showtime_id": input.ShowtimeID,
		"seat_id":     input.SeatID,
	})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *CartRepository) DeleteByID(ctx context.Context, ID int64) error {
	sql := `delete from public.carts where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *CartRepository) FindOne(ctx context.Context, filter CartFilter) (*Cart, error) {
	carts, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(carts) == 0 {
		return nil, NewErr(ErrNotFound, nil, "cart not found")
	}
	return &carts[0], nil
}

func (r *CartRepository) Find(ctx context.Context, filter CartFilter) ([]Cart, error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	sql := fmt.Sprintf(
		`
			select
				c.id,
				c.user_id,
				c.showtime_id,
				c.seat_id,
				c.created_at,
				c.updated_at,
				m.title as movie,
				s.start_at as showtime_start,
				s.end_at as showtime_end,
				r.name as room,
				st."name" as seat,
				s.price + st.additional_price as price
			from
				public.carts c
				join showtimes s on s.id  = c.showtime_id
				join movies m on m.id = s.movie_id
				join seats st on st.id = c.seat_id
				join rooms r on r.id = st.room_id
			where
				c.id in (%s)
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var carts []Cart
	for rows.Next() {
		var cart Cart
		err := rows.Scan(
			&cart.ID,
			&cart.UserID,
			&cart.ShowtimeID,
			&cart.SeatID,
			&cart.CreatedAt,
			&cart.UpdatedAt,
			&cart.Movie,
			&cart.ShowtimeStart,
			&cart.ShowtimeEnd,
			&cart.Room,
			&cart.Seat,
			&cart.Price,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		carts = append(carts, cart)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return carts, nil
}

func (r *CartRepository) Pagination(ctx context.Context, filter CartFilter, page PaginateInput) (*Paginate[Cart], error) {

	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf(`select count(*) from (%s)`, filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Cart{}, totalItems, page.Page, page.Size)
	if totalItems == 0 {
		return p, nil
	}

	if page.Page > p.TotalPage {
		page.Page = p.TotalPage
		p.CurrentPage = page.Page
	}

	sql = fmt.Sprintf(
		`
			select
				c.id,
				c.user_id,
				c.showtime_id,
				c.seat_id,
				c.created_at,
				c.updated_at,
				m.title as movie,
				s.start_at as showtime_start,
				s.end_at as showtime_end,
				r.name as room,
				st."name" as seat,
				s.price + st.additional_price as price
			from
				public.carts c
				join showtimes s on s.id  = c.showtime_id
				join movies m on m.id = s.movie_id
				join seats st on st.id = c.seat_id
				join rooms r on r.id = st.room_id
			where
				c.id in (%s)
			limit @page_size offset (@page - 1) * @page_size
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

	var carts []Cart
	for rows.Next() {
		var cart Cart
		err := rows.Scan(
			&cart.ID,
			&cart.UserID,
			&cart.ShowtimeID,
			&cart.SeatID,
			&cart.CreatedAt,
			&cart.UpdatedAt,
			&cart.Movie,
			&cart.ShowtimeStart,
			&cart.ShowtimeEnd,
			&cart.Room,
			&cart.Seat,
			&cart.Price,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		carts = append(carts, cart)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p.Items = carts
	return p, nil
}

func (r *CartRepository) getFilterSQL(_ context.Context, filter CartFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _c.id
		from public.carts _c
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_c.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_user_ids::int[], 1) > 0 then
					_c.user_id = any(@_user_ids)
				else
					true
			end
			and
			case
				when array_length(@_showtime_ids::int[], 1) > 0 then
					_c.showtime_id = any(@_showtime_ids)
				else
					true
			end
			and
			case
				when array_length(@_seat_ids::int[], 1) > 0 then
					_c.seat_id = any(@_seat_ids)
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":          filter.IDs,
		"_user_ids":     filter.UserIDs,
		"_showtime_ids": filter.ShowtimeIDs,
		"_seat_ids":     filter.SeatIDs,
	}
	return sql, args
}
