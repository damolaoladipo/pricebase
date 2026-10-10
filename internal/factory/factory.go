package factory

import "github.com/jackc/pgx/v5/pgxpool"

func Factories(db *pgxpool.Pool) {
	InitServices(db)
}
