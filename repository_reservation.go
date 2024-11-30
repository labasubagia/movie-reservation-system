package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewReservationRepository(tx pgx.Tx) *ReservationRepository {
	return &ReservationRepository{
		tx: tx,
	}
}

type ReservationRepository struct {
	tx pgx.Tx
}

func (r *ReservationRepository) Create(ctx context.Context, reservation *Reservation) (int64, error) {
	sql := `
		insert into public.reservations (user_id, total_price)
		values (@user_id, @total_price)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{
		"user_id":     reservation.UserID,
		"total_price": reservation.TotalPrice,
	}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *ReservationRepository) CreateItem(ctx context.Context, item *ReservationItem) (int64, error) {
	sql := `
		insert into public.reservation_items (user_id, showtime_id, seat_id, reservation_id, total_price)
		values (@user_id, @showtime_id, @seat_id, @reservation_id, @total_price)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{
		"user_id":        item.UserID,
		"showtime_id":    item.ShowtimeID,
		"seat_id":        item.SeatID,
		"reservation_id": item.ReservationID,
		"total_price":    item.TotalPrice,
	}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *ReservationRepository) UpdateByID(ctx context.Context, ID int64, input ReservationInput) error {
	sql := `
		update public.reservations
		set updated_at=now(), user_id=@user_id, total_price=@total_price, status=@status::public.reservation_status
		where id=@id
	`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":          ID,
		"user_id":     input.UserID,
		"status":      input.Status,
		"total_price": input.TotalPrice,
	})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *ReservationRepository) DeleteByID(ctx context.Context, ID int64) error {
	sql := `delete from public.reservations where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *ReservationRepository) FindOne(ctx context.Context, filter ReservationFilter) (*Reservation, error) {
	reservations, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(reservations) == 0 {
		return nil, NewErr(ErrNotFound, nil, "reservation not found")
	}
	return &reservations[0], nil
}

func (r *ReservationRepository) Find(ctx context.Context, filter ReservationFilter) ([]Reservation, error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	sql := fmt.Sprintf(
		`
			select
				r.id,
				r.user_id,
				r.status,
				r.total_price,
				r.created_at,
				r.updated_at
			from
				public.reservations r
			where
				r.id in (%s)
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var reservations []Reservation
	var reservationIDs []int64
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.Status,
			&reservation.TotalPrice,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		reservations = append(reservations, reservation)
		reservationIDs = append(reservationIDs, reservation.ID)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	if !filter.WithItems {
		return reservations, nil
	}

	items, err := r.FindItem(ctx, ReservationItemFilter{ReservationIDs: reservationIDs})
	if err != nil {
		return nil, err
	}
	// map[reservation_id]items
	mapReservationItems := map[int64][]ReservationItem{}
	for _, item := range items {
		mapReservationItems[item.ReservationID] = append(mapReservationItems[item.ReservationID], item)
	}

	for i, reservation := range reservations {
		reservations[i].Items = mapReservationItems[reservation.ID]
	}

	return reservations, nil
}

func (r *ReservationRepository) Pagination(ctx context.Context, filter ReservationFilter, page PaginateInput) (*Paginate[Reservation], error) {

	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf(`select count(*) from (%s)`, filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Reservation{}, totalItems, page.Page, page.Size)
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
				r.id,
				r.user_id,
				r.status,
				r.total_price,
				r.created_at,
				r.updated_at
			from
				public.reservations r
			where
				r.id in (%s)
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

	var reservations []Reservation
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.Status,
			&reservation.TotalPrice,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		reservations = append(reservations, reservation)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p.Items = reservations
	return p, nil
}

func (r *ReservationRepository) getFilterSQL(_ context.Context, filter ReservationFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _r.id
		from public.reservations _r
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_r.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_user_ids::int[], 1) > 0 then
					_r.user_id = any(@_user_ids)
				else
					true
			end
			and
			case
				when array_length(@_statuses::public.reservation_status[], 1) > 0 then
					_r.status = any(@_statuses::public.reservation_status[])
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":      filter.IDs,
		"_user_ids": filter.UserIDs,
		"_statuses": filter.Statuses,
	}
	return sql, args
}

func (r *ReservationRepository) FindItem(ctx context.Context, filter ReservationItemFilter) ([]ReservationItem, error) {
	filterSQL, filterArgs := r.getFilterItemSQL(ctx, filter)

	sql := fmt.Sprintf(
		`
			select
				rvi.id,
				rvi.user_id,
				rvi.reservation_id,
				rvi.showtime_id,
				rvi.seat_id,
				rvi.total_price,
				rvi.created_at,
				rvi.updated_at,
				m.title as movie,
				s.start_at as showtime_start,
				s.end_at as showtime_end,
				r.name as room,
				st."name" as seat
			from
				public.reservation_items rvi
				join showtimes s on s.id  = rvi.showtime_id
				join movies m on m.id = s.movie_id
				join seats st on st.id = rvi.seat_id
				join rooms r on r.id = st.room_id
			where
				rvi.id in (%s)
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var items []ReservationItem
	for rows.Next() {
		var item ReservationItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ReservationID,
			&item.ShowtimeID,
			&item.SeatID,
			&item.TotalPrice,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Movie,
			&item.ShowtimeStart,
			&item.ShowtimeEnd,
			&item.Room,
			&item.Seat,
		)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return items, nil
}

func (r *ReservationRepository) getFilterItemSQL(_ context.Context, filter ReservationItemFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _rvi.id
		from public.reservation_items _rvi
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_rvi.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_user_ids::int[], 1) > 0 then
					_rvi.user_id = any(@_user_ids)
				else
					true
			end
			and
			case
				when array_length(@_reservation_ids::int[], 1) > 0 then
					_rvi.reservation_id = any(@_reservation_ids)
				else
					true
			end
			and
			case
				when array_length(@_showtime_ids::int[], 1) > 0 then
					_rvi.showtime_id = any(@_showtime_ids)
				else
					true
			end
			and
			case
				when array_length(@_seat_ids::int[], 1) > 0 then
					_rvi.seat_id = any(@_seat_ids)
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":             filter.IDs,
		"_user_ids":        filter.UserIDs,
		"_reservation_ids": filter.ReservationIDs,
		"_showtime_ids":    filter.ShowtimeIDs,
		"_seat_ids":        filter.SeatIDs,
	}
	return sql, args
}
