package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewUserRepository(tx pgx.Tx) *UserRepository {
	return &UserRepository{
		tx: tx,
	}
}

type UserRepository struct {
	tx pgx.Tx
}

func (r *UserRepository) Create(ctx context.Context, user *User) (int64, error) {
	sql := `
		insert into public.users (email, password, role_id)
		select @email, @password, id from public.roles where name = @role
		returning id
	`
	var ID int64
	err := r.tx.QueryRow(ctx, sql, pgx.NamedArgs{
		"email":    user.Email,
		"password": user.PasswordHash,
		"role":     UserRegular,
	}).Scan(&ID)
	if err != nil {
		return 0, NewSQLErr(err)
	}
	return ID, nil
}

func (r *UserRepository) UpdatePasswordByID(ctx context.Context, ID int64, user *User) error {
	sql := `update public.users set updated_at=now(), password=@password where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID, "password": user.PasswordHash})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *UserRepository) UpdateRoleByID(ctx context.Context, ID int64, input UserInput) error {
	sql := `update public.users set updated_at=now(), role_id=@role_id where id=@id`
	_, err := r.tx.Exec(ctx, sql, pgx.NamedArgs{"id": ID, "role_id": input.RoleID})
	if err != nil {
		return NewSQLErr(err)
	}
	return nil
}

func (r *UserRepository) FindOne(ctx context.Context, filter UserFilter) (*User, error) {
	users, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, NewErr(ErrNotFound, nil, "user not found")
	}
	return &users[0], nil
}

func (r *UserRepository) Find(ctx context.Context, filter UserFilter) ([]User, error) {
	filterSQL, filterArgs := r.getFilterSQL(ctx, filter)

	sql := fmt.Sprintf(
		`
			select
				u.id,
				u.email,
				u.password,
				u.created_at,
				u.updated_at,
				r.name
			from
				public.users u
			join public.roles r on
				u.role_id = r.id
			where
				u.id in (%s)
		`,
		filterSQL,
	)
	rows, err := r.tx.Query(ctx, sql, filterArgs)
	if err != nil {
		return nil, NewSQLErr(err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &user.Role)
		if err != nil {
			return nil, NewSQLErr(err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, NewSQLErr(err)
	}

	return users, nil
}

func (r *UserRepository) getFilterSQL(_ context.Context, filter UserFilter) (sql string, args pgx.NamedArgs) {
	sql = `
		select distinct _u.id
		from public.users _u
		join public.roles _r on _u.role_id = _r.id
		where
			case
				when array_length(@_ids::int[], 1) > 0 then
					_u.id = any(@_ids)
				else
					true
			end
			and
			case
				when array_length(@_emails::text[], 1) > 0 then
					_u.email = any(@_emails)
				else
					true
			end
			and
			case
				when array_length(@_role_ids::int[], 1) > 0 then
					_u.role_id = any(@_role_ids)
				else
					true
			end
			and
			case
				when array_length(@_roles::text[], 1) > 0 then
					_r.name = any(@_roles)
				else
					true
			end
	`
	args = pgx.NamedArgs{
		"_ids":      filter.IDs,
		"_emails":   filter.Emails,
		"_role_ids": filter.RoleIDs,
		"_roles":    filter.Roles,
	}
	return sql, args
}
