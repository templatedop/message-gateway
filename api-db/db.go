package db

import (
	"context"
	"errors"
	"fmt"

	l "MgApplication/api-log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/**
 * DB is a wrapper for PostgreSQL database connection
 * that uses pgxpool as database driver
 */

type DB struct {
	*pgxpool.Pool
}

type DBInterface interface {
	Close()
	WithTx(ctx context.Context, fn func(tx pgx.Tx) error, levels ...pgx.TxIsoLevel) error
	ReadTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

var _ DBInterface = (*DB)(nil)

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) Ping() error {
	return db.Pool.Ping(context.Background())
}

func (db *DB) PingContext(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) WithTx(ctx context.Context, fn func(tx pgx.Tx) error, levels ...pgx.TxIsoLevel) error {
	var level pgx.TxIsoLevel
	if len(levels) > 0 {
		level = levels[0]
	} else {
		level = pgx.ReadCommitted // Default value
	}

	l.Debug(ctx, "level set to "+level)
	return db.inTx(ctx, level, "", fn)
}

func (db *DB) ReadTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return db.inTx(ctx, pgx.ReadCommitted, pgx.ReadOnly, fn)
}

func (db *DB) inTx(ctx context.Context, level pgx.TxIsoLevel, access pgx.TxAccessMode,
	fn func(tx pgx.Tx) error) (err error) {

	conn, errAcq := db.Pool.Acquire(ctx)
	if errAcq != nil {
		return fmt.Errorf("acquiring connection: %w", errAcq)
	}
	defer conn.Release()

	opts := pgx.TxOptions{
		IsoLevel:   level,
		AccessMode: access,
	}

	tx, errBegin := conn.BeginTx(ctx, opts)
	if errBegin != nil {
		return fmt.Errorf("begin tx: %w", errBegin)
	}

	defer func() {
		errRollback := tx.Rollback(ctx)
		if !(errRollback == nil || errors.Is(errRollback, pgx.ErrTxClosed)) {
			err = errRollback
		}
	}()

	if err := fn(tx); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return fmt.Errorf("rollback tx: %v (original: %w)", errRollback, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
