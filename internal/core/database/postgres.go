package database

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"taskTestEffectMobile/internal/core/configs"
)

func RunMigrations(dbUrl string) error {
	m, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			panic(err)
		}
	}(m)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func CreateDBConnection() (*sql.DB, error) {
	cfg := configs.Init()
	dbUrl := cfg.DB.DBUrl()
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
