package repository

import (
	"context"
	errs "github.com/cristiancll/go-errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qrpay-wpp/internal/errCode"
)

type TCreater[E any] interface {
	TCreate(context.Context, pgx.Tx, *E) error
}

func tCreate(ctx context.Context, tx pgx.Tx, query string, args ...any) (int64, error) {
	row := tx.QueryRow(ctx, query, args...)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		return id, errs.New(err, errCode.Internal)
	}
	return id, nil
}

type TUpdater[E any] interface {
	TUpdate(context.Context, pgx.Tx, *E) error
}

func tUpdate(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	cmd, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	if cmd.RowsAffected() == 0 {
		return errs.New(err, errCode.NotChanged)
	}
	return nil
}

type TDeleter[E any] interface {
	TDelete(context.Context, pgx.Tx, *E) error
}

func tDelete(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	cmd, err := tx.Exec(ctx, query, args)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	if cmd.RowsAffected() == 0 {
		return errs.New(err, errCode.NotChanged)
	}
	return nil
}

type TGetterById[E any] interface {
	TGetById(ctx context.Context, tx pgx.Tx, id int64) (*E, error)
}
type TGetterByUUID[E any] interface {
	TGetByUUID(ctx context.Context, tx pgx.Tx, uuid string) (*E, error)
}

func tGet[T any](ctx context.Context, tx pgx.Tx, query string, args ...any) (*T, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.New(err, errCode.Internal)
	}
	defer rows.Close()
	entity, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.New(err, errCode.NotChanged)
		}
		return nil, errs.New(err, errCode.Internal)
	}
	return &entity, nil
}

type TGetterAll[E any] interface {
	TGetAll(context.Context, pgx.Tx) ([]*E, error)
}

func tGetAll[T any](ctx context.Context, tx pgx.Tx, query string, args ...any) ([]*T, error) {
	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, errs.New(err, errCode.Internal)
	}
	defer rows.Close()
	entities, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.New(err, errCode.NotChanged)
		}
		return nil, errs.New(err, errCode.Internal)
	}
	entitiesPtr := make([]*T, len(entities))
	for i := range entities {
		entitiesPtr[i] = &entities[i]
	}
	return entitiesPtr, nil
}

type TCounter[E any] interface {
	TCount(context.Context, pgx.Tx, ...any) (int64, error)
}

func tCount(ctx context.Context, tx pgx.Tx, query string, args ...any) (int64, error) {
	row := tx.QueryRow(ctx, query, args...)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		if err == pgx.ErrNoRows {
			return count, errs.New(err, errCode.NotChanged)
		}
		return count, errs.New(err, errCode.Internal)
	}
	return count, nil
}

type Migrater interface {
	Migrate(context.Context) error
}

func migrate(ctx context.Context, pool *pgxpool.Pool, query string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	return nil
}

type TCRUDer[E any] interface {
	TCreater[E]
	TUpdater[E]
	TDeleter[E]
	TGetterById[E]
	TGetterByUUID[E]
	TGetterAll[E]
}
