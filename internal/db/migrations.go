package db

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, dsn string) error {
    sqldb, err := sql.Open("pgx", dsn)
    if err != nil {
        return fmt.Errorf("open sql: %w", err)
    }
    defer sqldb.Close()

    if err := sqldb.PingContext(ctx); err != nil {
        return fmt.Errorf("ping sql: %w", err)
    }

    goose.SetDialect("postgres")
    if err := goose.UpContext(ctx, sqldb, "./migrations"); err != nil {
        return fmt.Errorf("migrations up: %w", err)
    }
    return nil
}


