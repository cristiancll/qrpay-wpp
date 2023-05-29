package model

import "time"

type WhatsApp struct {
	ID        int64     `db:"id"`
	UUID      string    `db:"uuid"`
	QR        string    `db:"qr"`
	Phone     *string   `db:"phone"`
	Scanned   bool      `db:"scanned"`
	Active    bool      `db:"active"`
	Banned    bool      `db:"banned"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
