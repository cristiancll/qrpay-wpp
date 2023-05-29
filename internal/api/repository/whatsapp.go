package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qrpay-wpp/internal/api/model"
	"time"
)

type WhatsApp interface {
	Migrater
	TCreater[model.WhatsApp]
	TGetterByUUID[model.WhatsApp]
	TGetterAll[model.WhatsApp]
}

type whatsApp struct {
	db *pgxpool.Pool
}

func NewWhatsApp(db *pgxpool.Pool) WhatsApp {
	return &whatsApp{db: db}
}

func (r *whatsApp) GetByPhone(ctx context.Context, tx pgx.Tx, phone string) (*model.WhatsApp, error) {
	query := `SELECT id, uuid, qr, phone, scanned, active, banned, created_at, updated_at FROM whatsapps WHERE phone = $1`
	return tGet[model.WhatsApp](ctx, tx, query, phone)
}

func (r *whatsApp) ClearUnusedWhatsApp(ctx context.Context, tx pgx.Tx) error {
	query := `DELETE FROM whatsapps WHERE scanned = FALSE AND active = FALSE`
	return tDelete(ctx, tx, query)
}

func (r *whatsApp) TGetUnscannedWhatsApp(ctx context.Context, tx pgx.Tx) (*model.WhatsApp, error) {
	query := `SELECT id, uuid, qr, phone, scanned, active, banned, created_at, updated_at FROM whatsapps WHERE scanned = FALSE LIMIT 1`
	return tGet[model.WhatsApp](ctx, tx, query)
}

func (r *whatsApp) TGetActiveWhatsApp(ctx context.Context, tx pgx.Tx) (*model.WhatsApp, error) {
	query := `SELECT id, uuid, qr, phone, scanned, active, banned, created_at, updated_at FROM whatsapps WHERE active = TRUE LIMIT 1`
	return tGet[model.WhatsApp](ctx, tx, query)
}

func (r *whatsApp) DisableAll(ctx context.Context, tx pgx.Tx) error {
	query := `UPDATE whatsapps SET active = FALSE, updated_at = $1 WHERE active = TRUE`
	return tUpdate(ctx, tx, query, time.Now().UTC())
}

func (r *whatsApp) TCreate(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	whats.UUID = uuid.New().String()
	whats.CreatedAt = time.Now().UTC()
	whats.UpdatedAt = time.Now().UTC()
	query := `INSERT INTO whatsapps (uuid, qr, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`
	id, err := tCreate(ctx, tx, query, whats.UUID, whats.QR, whats.CreatedAt, whats.UpdatedAt)
	if err != nil {
		return err
	}
	whats.ID = id
	return nil
}

func (r *whatsApp) TGetByUUID(ctx context.Context, tx pgx.Tx, uuid string) (*model.WhatsApp, error) {
	query := `SELECT id, uuid, qr, phone, scanned, active, banned, created_at, updated_at FROM whatsapps WHERE uuid = $1`
	return tGet[model.WhatsApp](ctx, tx, query, uuid)
}

func (r *whatsApp) TGetAll(ctx context.Context, tx pgx.Tx) ([]*model.WhatsApp, error) {
	query := `SELECT id, uuid, qr, phone, scanned, active, banned, created_at, updated_at FROM whatsapps`
	return tGetAll[model.WhatsApp](ctx, tx, query)
}

func (r *whatsApp) TUpdate(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	whats.UpdatedAt = time.Now().UTC()
	query := `UPDATE whatsapps SET qr = $2, phone = $3, scanned = $4, active = $5, banned = $6, updated_at = $7 WHERE id = $1`
	return tUpdate(ctx, tx, query, whats.ID, whats.QR, whats.Phone, whats.Scanned, whats.Active, whats.Banned, whats.UpdatedAt)
}

func (r *whatsApp) TDelete(ctx context.Context, tx pgx.Tx, whats *model.WhatsApp) error {
	query := `DELETE FROM whatsapps WHERE id = $1`
	return tDelete(ctx, tx, query, whats.ID)
}

func (r *whatsApp) TCountByQRCode(ctx context.Context, tx pgx.Tx, qrCode string) (int64, error) {
	query := `SELECT COUNT(*) FROM whatsapps WHERE qr = $1`
	return tCount(ctx, tx, query, qrCode)
}

func (r *whatsApp) Migrate(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS whatsapps (
				id SERIAL PRIMARY KEY, 
				uuid VARCHAR(255) NOT NULL, 
				qr VARCHAR(255) NOT NULL, 
				phone VARCHAR(255), 
				active BOOLEAN DEFAULT FALSE, 
				scanned BOOLEAN DEFAULT FALSE,
				banned BOOLEAN DEFAULT FALSE, 
				created_at TIMESTAMP NOT NULL, 
				updated_at TIMESTAMP NOT NULL
			)`
	return migrate(ctx, r.db, query)
}
