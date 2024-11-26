-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.roles (
    id serial NOT NULL,
    "name" varchar NOT NULL,
    CONSTRAINT roles_pk PRIMARY KEY (id),
    CONSTRAINT roles_unique UNIQUE ("name")
);
-- seed
INSERT INTO public.roles ("name") VALUES ('user');
INSERT INTO public.roles ("name") VALUES ('admin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.roles;
-- +goose StatementEnd
