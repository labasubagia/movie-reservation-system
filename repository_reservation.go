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
		insert into public.reservations (user_id, showtime_id, seat_id, status)
		values (@user_id, @showtime_id, @seat_id, @status::public.reservation_status)
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{
		"user_id":     reservation.UserID,
		"showtime_id": reservation.ShowtimeID,
		"seat_id":     reservation.SeatID,
		"status":      reservation.Status,
	}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *ReservationRepository) UpdateByID(ctx context.Context, ID int64, input ReservationInput) error {
	sql := `
		update public.reservations
		set updated_at=now(), user_id=@user_id, showtime_id=@showtime_id, seat_id=@seat_id, status=@status::public.reservation_status
		where id=@id
	`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":          ID,
		"user_id":     input.UserID,
		"showtime_id": input.ShowtimeID,
		"seat_id":     input.SeatID,
		"status":      input.Status,
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
				r.showtime_id,
				r.seat_id,
				r.status,
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
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ShowtimeID,
			&reservation.SeatID,
			&reservation.Status,
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
				r.showtime_id,
				r.seat_id,
				r.status,
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
			&reservation.ShowtimeID,
			&reservation.SeatID,
			&reservation.Status,
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
				when array_length(@_showtime_ids::int[], 1) > 0 then
					_r.showtime_id = any(@_showtime_ids)
				else
					true
			end
			and
			case
				when array_length(@_seat_ids::int[], 1) > 0 then
					_r.seat_id = any(@_seat_ids)
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
		"_ids":          filter.IDs,
		"_user_ids":     filter.UserIDs,
		"_showtime_ids": filter.ShowtimeIDs,
		"_seat_ids":     filter.SeatIDs,
		"_statuses":     filter.Statuses,
	}
	return sql, args
}
