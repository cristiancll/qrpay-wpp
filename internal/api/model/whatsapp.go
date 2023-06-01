package model

import "time"

type WhatsApp struct {
	ID          int64     `db:"id"`
	UUID        string    `db:"uuid"`
	AccountUUID string    `db:"account_uuid"`
	Phone       string    `db:"phone"`
	Connected   bool      `db:"connected"`
	Active      bool      `db:"active"`
	Banned      bool      `db:"banned"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
