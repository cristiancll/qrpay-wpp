package repository

import (
	"context"
	errs "github.com/cristiancll/go-errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qrpay-wpp/internal/api/model"
	"time"
)

type WhatsApp interface {
	Migrater
	TCreater[model.WhatsApp]
	TUpdater[model.WhatsApp]
	TGetterByUUID[model.WhatsApp]
	TGetterAll[model.WhatsApp]
	TGetByAccountId(ctx context.Context, tx pgx.Tx, accountUUID string) (*model.WhatsApp, error)
}

type whatsApp struct {
	db *pgxpool.Pool
}

func NewWhatsApp(db *pgxpool.Pool) WhatsApp {
	return &whatsApp{db: db}
}

func (r *whatsApp) TGetByAccountId(ctx context.Context, tx pgx.Tx, accountUUID string) (*model.WhatsApp, error) {
	query := `SELECT * FROM whatsapps WHERE account_uuid = $1 AND active = true`
	return tGet[model.WhatsApp](ctx, tx, query, accountUUID)
}

func (r *whatsApp) TCreate(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	whats.UUID = uuid.New().String()
	whats.CreatedAt = time.Now().UTC()
	whats.UpdatedAt = time.Now().UTC()
	query := `INSERT INTO whatsapps (uuid, account_uuid, phone, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	id, err := tCreate(ctx, tx, query, whats.UUID, whats.AccountUUID, whats.Phone, whats.CreatedAt, whats.UpdatedAt)
	if err != nil {
		return errs.Wrap(err, "")
	}
	whats.ID = id
	return nil
}

func (r *whatsApp) TGetByUUID(ctx context.Context, tx pgx.Tx, uuid string) (*model.WhatsApp, error) {
	query := `SELECT * FROM whatsapps WHERE uuid = $1`
	return tGet[model.WhatsApp](ctx, tx, query, uuid)
}

func (r *whatsApp) TGetAll(ctx context.Context, tx pgx.Tx) ([]*model.WhatsApp, error) {
	query := `SELECT * FROM whatsapps`
	return tGetAll[model.WhatsApp](ctx, tx, query)
}

func (r *whatsApp) TUpdate(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	whats.UpdatedAt = time.Now().UTC()
	query := `UPDATE whatsapps SET connected = $2, active = $3, banned = $4, updated_at = $5 WHERE id = $1`
	return tUpdate(ctx, tx, query, whats.ID, whats.Connected, whats.Active, whats.Banned, whats.UpdatedAt)
}

func (r *whatsApp) TDelete(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	query := `DELETE FROM whatsapps WHERE id = $1`
	return tDelete(ctx, tx, query, whats.ID)
}

func (r *whatsApp) Migrate(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS whatsapps (
				id SERIAL PRIMARY KEY, 
				uuid VARCHAR(255) NOT NULL, 
				account_uuid VARCHAR(255) NOT NULL,
				phone VARCHAR(255) NOT NULL, 
				connected BOOLEAN DEFAULT FALSE,
				active BOOLEAN DEFAULT FALSE, 
				banned BOOLEAN DEFAULT FALSE, 
				created_at TIMESTAMP NOT NULL, 
				updated_at TIMESTAMP NOT NULL
			)`
	return migrate(ctx, r.db, query)
}
