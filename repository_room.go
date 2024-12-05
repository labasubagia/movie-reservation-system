package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewRoomRepository(tx pgx.Tx) *RoomRepository {
	return &RoomRepository{
		tx: tx,
	}
}

type RoomRepository struct {
	tx pgx.Tx
}

func (r *RoomRepository) Create(ctx context.Context, room *Room) (int64, error) {
	sql := `insert into public.rooms (name) values (@name) returning id`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{"name": room.Name}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *RoomRepository) UpdateByID(ctx context.Context, ID int64, input RoomInput) error {
	sql := `update public.rooms SET updated_at=now(), name=@name where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{
		"id":   ID,
		"name": input.Name,
	})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *RoomRepository) DeleteByID(ctx context.Context, ID int64) error {
	sql := `delete from public.rooms where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *RoomRepository) FindOne(ctx context.Context, filter RoomFilter) (*Room, error) {
	rooms, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(rooms) == 0 {
		return nil, NewErr(ErrNotFound, nil, "room not found")
	}
	return &rooms[0], nil
}

func (r *RoomRepository) Find(ctx context.Context, filter RoomFilter) ([]Room, error) {

	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)
	sql := fmt.Sprintf(
		`
			select r.id, r.name, r.created_at, r.updated_at, count(s.id) as capacity
			from public.rooms r
			left join public.seats s on r.id = s.room_id
			where r.id in (%s)
			group by r.id
			order by capacity desc, r.name asc

		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.CreatedAt, &room.UpdatedAt, &room.Capacity)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		rooms = append(rooms, room)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return rooms, nil
}

func (r *RoomRepository) Pagination(ctx context.Context, filter RoomFilter, page PaginateInput) (*Paginate[Room], error) {

	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	var totalItems int64
	sql := fmt.Sprintf(`select count(*) from (%s)`, filterSQL)
	err := r.tx.QueryRow(ctx, sql, filterArgs).Scan(&totalItems)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p := NewPaginate([]Room{}, totalItems, page.Page, page.Size)
	if totalItems == 0 {
		return p, nil
	}

	if page.Page > p.TotalPage {
		page.Page = p.TotalPage
		p.CurrentPage = page.Page
	}

	sql = fmt.Sprintf(
		`
			select r.id, r.name, r.created_at, r.updated_at, count(s.*) as capacity
			from public.rooms r
			left join public.seats s on r.id = s.room_id
			where r.id in (%s)
			group by r.id
			order by capacity desc, r.name asc
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

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.CreatedAt, &room.UpdatedAt, &room.Capacity)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		rooms = append(rooms, room)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}
	p.Items = rooms
	return p, nil
}

func (r *RoomRepository) getFilterSQL(_ context.Context, filter RoomFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _r.id
		from public.rooms _r
		left join public.seats _s on _r.id = _s.room_id
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_r.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_names::text[], 1) > 0 then
					_r.name = any(@_names)
				else
					true
			end
		group by _r.id
		having
			case
				when @_is_usable::bool is not null then
					(count(_s.id) > 0) = @_is_usable
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":       filter.IDs,
		"_names":     filter.Names,
		"_is_usable": filter.IsUsable,
	}
	return sql, args
}

// seats

func (r *RoomRepository) SetSeats(ctx context.Context, roomID int64, inputs []Seat) (err error) {

	nameMap := map[string]struct{}{}
	for _, seat := range inputs {
		nameMap[seat.Name] = struct{}{}
	}

	oldSeats, err := r.FilterSeats(ctx, SeatFilter{
		RoomIDs: []int64{roomID},
	})
	if err != nil {
		return err
	}
	deletedIDs := []int64{}
	oldSeatMap := map[string]Seat{}
	for _, item := range oldSeats {
		oldSeatMap[item.Name] = item
		if _, ok := nameMap[item.Name]; !ok {
			deletedIDs = append(deletedIDs, item.ID)
		}
	}

	newRows := []Seat{}
	updateRows := []Seat{}
	for _, item := range inputs {
		oldSeat, ok := oldSeatMap[item.Name]
		if ok {
			if oldSeat.AdditionalPrice == item.AdditionalPrice {
				continue
			}
			oldSeat.AdditionalPrice = item.AdditionalPrice
			updateRows = append(updateRows, oldSeat)
			continue
		}
		newRows = append(newRows, item)
	}

	if len(deletedIDs) > 0 {
		err := r.BulkDeleteSeat(ctx, deletedIDs)
		if err != nil {
			return err
		}
	}

	if len(updateRows) > 0 {
		err = r.BulkUpdateSeat(ctx, newRows)
		if err != nil {
			return err
		}
	}

	if len(newRows) > 0 {
		err = r.BulkCreateSeat(ctx, newRows)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RoomRepository) FilterSeats(ctx context.Context, filter SeatFilter) ([]Seat, error) {
	filterSQL, filterArgs := r.getFilterSeatSQL(ctx, filter)

	sql := fmt.Sprintf(
		`
			select s.id, s.room_id, s.name, s.additional_price, s.created_at, s.updated_at
			from public.seats s
			where s.id in (%s)
			order by s.name asc, s.room_id asc, s.id asc
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
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
			&seat.Name,
			&seat.AdditionalPrice,
			&seat.CreatedAt,
			&seat.UpdatedAt,
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

func (r *RoomRepository) BulkCreateSeat(ctx context.Context, newSeats []Seat) error {
	if len(newSeats) == 0 {
		return NewErr(ErrInput, nil, "new seats empty")
	}

	newRows := make([][]any, 0, len(newSeats))
	for _, item := range newSeats {
		newRows = append(newRows, []any{item.RoomID, item.Name, item.AdditionalPrice})
	}

	_, err := r.tx.CopyFrom(
		ctx,
		pgx.Identifier{"seats"},
		[]string{"room_id", "name", "additional_price"},
		pgx.CopyFromRows(newRows),
	)
	if err != nil {
		return NewSQLErr(err)
	}

	return nil
}

func (r *RoomRepository) BulkUpdateSeat(ctx context.Context, updateSeats []Seat) error {
	if len(updateSeats) == 0 {
		return NewErr(ErrInput, nil, "update seats empty")
	}

	sql := `
		update public.seats
		set room_id=@room_id, "name"=@name, additional_price=@additional_price, updated_at=now()
		where id=@id
	`

	batch := pgx.Batch{}
	for _, seat := range updateSeats {
		batch.Queue(sql, pgx.NamedArgs{
			"id":               seat.ID,
			"name":             seat.Name,
			"room_id":          seat.RoomID,
			"additional_price": seat.AdditionalPrice,
		})
	}
	br := r.tx.SendBatch(ctx, &batch)

	for range updateSeats {
		_, err := br.Exec()
		if err != nil {
			br.Close()
			return NewSQLErr(err)
		}
	}

	if err := br.Close(); err != nil {
		return NewSQLErr(err)
	}

	return nil
}

func (r *RoomRepository) BulkDeleteSeat(ctx context.Context, IDs []int64) error {
	if len(IDs) == 0 {
		return NewErr(ErrInput, nil, "ids seats empty")
	}

	sql := `delete from public.seats where id = any(@ids)`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"ids": IDs})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *RoomRepository) getFilterSeatSQL(_ context.Context, filter SeatFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select _s.id
		from public.seats _s
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_s.id = any(@_ids)
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
				when array_length(@_names::text[], 1) > 0 then
					_s.name = any(@_names)
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":      filter.IDs,
		"_room_ids": filter.RoomIDs,
		"_names":    filter.Names,
	}
	return sql, args
}
