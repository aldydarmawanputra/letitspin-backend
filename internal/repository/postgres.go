package repository

import "database/sql"

type PostgresRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}
