package pgmigrate

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func DBConnect(dbDSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, err
	}
	//ensure the database is accessible
	err = db.Ping()
	if err != nil {
		return db, err
	}
	return db, nil
}

func ApplyMigrations(migrations_dir string, db *sql.DB) error {
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrations_dir,
		"postgres", dbDriver)
	if err != nil {
		return err
	}
	defer m.Close()
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
