-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.users (
	id bigserial NOT NULL,
	email varchar NOT NULL,
	"password" varchar NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	updated_at timestamptz DEFAULT now() NOT NULL,
	role_id int NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT users_unique UNIQUE (email),
	CONSTRAINT users_roles_fk FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- seed
-- pass: 12345678
INSERT INTO public.users (email, "password", role_id)
SELECT 'admin@gmail.com', '$2a$10$P.CkvHWScvCiHL1bCyVDfeYM0RahNZBx6grqw9qcRUJ9xZxzFlO56', id FROM public.roles WHERE "name" = 'admin';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.users;
-- +goose StatementEnd
