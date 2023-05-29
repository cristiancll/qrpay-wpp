package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"qrpay-wpp/internal/api/repository"
)

type WhatsApp interface {
}

type whatsApp struct {
	pool *pgxpool.Pool
	repo repository.WhatsApp
}

func NewWhatsApp(pool *pgxpool.Pool, repo repository.WhatsApp) WhatsApp {
	return &whatsApp{
		pool: pool,
		repo: repo,
	}
}
