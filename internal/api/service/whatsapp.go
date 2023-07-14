package service

import (
	"context"
	"fmt"
	errs "github.com/cristiancll/go-errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mau.fi/whatsmeow/types/events"
	"qrpay-wpp/internal/api/model"
	"qrpay-wpp/internal/api/repository"
	server "qrpay-wpp/internal/api/system"
	"qrpay-wpp/internal/errCode"
)

type WhatsApp interface {
	Connect(ctx context.Context, uuid string) error
	Message(ctx context.Context, uuid string, to string, text string, media []byte) error
	GetQRCode(uuid string) (string, error)
}

type whatsApp struct {
	pool   *pgxpool.Pool
	repo   repository.WhatsApp
	system server.WhatsAppSystem
}

func NewWhatsApp(pool *pgxpool.Pool, repo repository.WhatsApp, system server.WhatsAppSystem) WhatsApp {
	return &whatsApp{
		pool:   pool,
		repo:   repo,
		system: system,
	}
}

func (s *whatsApp) create(ctx context.Context, accountUUID string, phone string) (*model.WhatsApp, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errs.New(err, errCode.Internal)
	}
	defer tx.Rollback(ctx)

	wpp, err := s.repo.TGetByAccountId(ctx, tx, accountUUID)
	if err != nil {
		return nil, errs.Wrap(err, "")
	}

	wpp = &model.WhatsApp{
		AccountUUID: accountUUID,
		Phone:       phone,
	}
	err = s.repo.TCreate(ctx, tx, wpp)
	if err != nil {
		return nil, errs.Wrap(err, "")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, errs.New(err, errCode.Internal)
	}
	return wpp, nil
}

func (s *whatsApp) update(ctx context.Context, accountUUID string, isConnected, isActive, isBanned bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	defer tx.Rollback(ctx)
	wpp, err := s.repo.TGetByAccountId(ctx, tx, accountUUID)
	if err != nil {
		return errs.Wrap(err, "")
	}
	wpp.Active = isActive
	wpp.Connected = isConnected
	wpp.Banned = isBanned
	err = s.repo.TUpdate(ctx, tx, wpp)
	if err != nil {
		return errs.Wrap(err, "")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	return nil
}

func (s *whatsApp) handleUserResponse(ctx context.Context, accountUUID string, phone string, msg string) error {
	return nil
}

func (s *whatsApp) eventHandler(accountUUID string, evt any) {
	ctx := context.Background()
	switch v := evt.(type) {
	case *events.PairSuccess:
		phone := v.ID.User
		_, err := s.create(ctx, accountUUID, phone)
		if err != nil {
			// TODO: log error
			return
		}
	case *events.Connected:
		fmt.Printf("Connected: %+v\n", v)
		s.update(ctx, accountUUID, true, true, false)
	case *events.Disconnected:
		fmt.Printf("Disconnected: %+v\n", v)
		s.update(ctx, accountUUID, false, true, false)
	case *events.TemporaryBan:
		fmt.Printf("TemporaryBan: %+v\n", v)
		s.update(ctx, accountUUID, false, false, true)
	case *events.LoggedOut:
		fmt.Printf("LoggedOut: %+v\n", v)
		s.update(ctx, accountUUID, false, false, false)
	case *events.Message:
		fmt.Printf("Message: %+v\n", v)
		phone := v.Info.MessageSource.Sender.User
		msg := *v.Message.Conversation
		s.handleUserResponse(ctx, accountUUID, phone, msg)
	}

}

func (s *whatsApp) Connect(ctx context.Context, uuid string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	defer tx.Rollback(ctx)
	wpp, _ := s.repo.TGetByAccountId(ctx, tx, uuid)
	if wpp != nil {
		err = s.system.Connect(ctx, wpp.AccountUUID, wpp.Phone, s.eventHandler)
		if err != nil {
			return errs.Wrap(err, "")
		}
	} else {
		err = s.system.Connect(ctx, uuid, "", s.eventHandler)
		if err != nil {
			return errs.Wrap(err, "")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	return nil
}

func (s *whatsApp) Message(ctx context.Context, uuid string, to string, text string, media []byte) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	defer tx.Rollback(ctx)
	wpp, err := s.repo.TGetByAccountId(ctx, tx, uuid)
	if err != nil {
		return errs.Wrap(err, "")
	}
	err = s.system.SendMessage(ctx, wpp.AccountUUID, to, text, media)
	if err != nil {
		return errs.Wrap(err, "")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.New(err, errCode.Internal)
	}
	return nil
}

func (s *whatsApp) GetQRCode(uuid string) (string, error) {
	return s.system.GetQRCode(uuid)
}
