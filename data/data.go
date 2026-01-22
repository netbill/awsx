package data

import (
	"context"
	"database/sql"

	"github.com/netbill/imgx/data/sqlc"
	"github.com/netbill/pgx"
)

type Data struct {
	db *sql.DB
}

func (d Data) Queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(pgx.Exec(d.db, ctx))
}

func (d Data) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pgx.Transaction(d.db, ctx, fn)
}
