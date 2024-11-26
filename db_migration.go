package main

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed migration/*.sql
var embedMigrations embed.FS

var MIGRATE_VERSION int64 = 20241125035656

func migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}
	if version == MIGRATE_VERSION {
		return nil
	}
	if version > MIGRATE_VERSION {
		err := goose.DownTo(db, "migration", MIGRATE_VERSION)
		if err != nil {
			return err
		}
		return nil
	}
	err = goose.UpTo(db, "migration", MIGRATE_VERSION)
	if err != nil {
		return err
	}
	return nil
}
