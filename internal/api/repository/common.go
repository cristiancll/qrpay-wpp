package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"qrpay-wpp/internal/errors"
)

type TCreater[E any] interface {
	TCreate(context.Context, pgx.Tx, *E) error
}

func tCreate(ctx context.Context, tx pgx.Tx, query string, args ...any) (int64, error) {
	row := tx.QueryRow(ctx, query, args...)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		return id, status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	return id, nil
}

type TUpdater[E any] interface {
	TUpdate(context.Context, pgx.Tx, *E) error
}

func tUpdate(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	cmd, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	if cmd.RowsAffected() == 0 {
		return status.Error(codes.NotFound, errors.NO_ROWS_AFFECTED)
	}
	return nil
}

type TDeleter[E any] interface {
	TDelete(context.Context, pgx.Tx, *E) error
}

func tDelete(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	cmd, err := tx.Exec(ctx, query, args)
	if err != nil {
		return status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	if cmd.RowsAffected() == 0 {
		return status.Error(codes.NotFound, errors.NO_ROWS_AFFECTED)
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
		return nil, status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	defer rows.Close()
	entity, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, errors.NO_ROWS_FOUND)
		}
		return nil, status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	return &entity, nil
}

type TGetterAll[E any] interface {
	TGetAll(context.Context, pgx.Tx) ([]*E, error)
}

func tGetAll[T any](ctx context.Context, tx pgx.Tx, query string, args ...any) ([]*T, error) {
	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	defer rows.Close()
	entities, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, errors.NO_ROWS_FOUND)
		}
		return nil, status.Error(codes.Internal, errors.DATABASE_ERROR)
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
			return count, status.Error(codes.NotFound, errors.NO_ROWS_FOUND)
		}
		return count, status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	return count, nil
}

type Migrater interface {
	Migrate(context.Context) error
}

func migrate(ctx context.Context, pool *pgxpool.Pool, query string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return status.Error(codes.Internal, errors.DATABASE_ERROR)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return status.Error(codes.Internal, errors.DATABASE_ERROR)
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
