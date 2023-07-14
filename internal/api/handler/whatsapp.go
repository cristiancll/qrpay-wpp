package handler

import (
	"context"
	"errors"
	errs "github.com/cristiancll/go-errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "qrpay-wpp/internal/api/proto/generated"
	"qrpay-wpp/internal/api/service"
	"qrpay-wpp/internal/errCode"
	"time"
)

type WhatsApp interface {
	Connect(ctx context.Context, req *proto.WhatsAppConnectRequest) (*proto.WhatsAppConnectResponse, error)
	Message(ctx context.Context, req *proto.WhatsAppMessageRequest) (*proto.WhatsAppMessageResponse, error)
	QRCode(req *proto.WhatsAppQRRequest, stream proto.WhatsAppService_QRServer) error
	proto.WhatsAppServiceServer
}

type whatsApp struct {
	service service.WhatsApp
	proto.UnimplementedWhatsAppServiceServer
}

func NewWhatsApp(s service.WhatsApp) WhatsApp {
	return &whatsApp{service: s}
}

func (h *whatsApp) Connect(ctx context.Context, req *proto.WhatsAppConnectRequest) (*proto.WhatsAppConnectResponse, error) {
	if req.AccountUUID == "" {
		return nil, errs.New(errors.New(""), errCode.InvalidArgument)
	}
	err := h.service.Connect(ctx, req.AccountUUID)
	if err != nil {
		return nil, errs.Wrap(err, "")
	}
	return &proto.WhatsAppConnectResponse{}, nil
}

func (h *whatsApp) Message(ctx context.Context, req *proto.WhatsAppMessageRequest) (*proto.WhatsAppMessageResponse, error) {
	if req.AccountUUID == "" {
		return nil, errs.New(errors.New(""), errCode.InvalidArgument)
	}
	if req.To == "" {
		return nil, errs.New(errors.New(""), errCode.InvalidArgument)
	}
	if req.Text == "" || req.Media == nil {
		return nil, errs.New(errors.New(""), errCode.InvalidArgument)
	}
	err := h.service.Message(ctx, req.AccountUUID, req.To, req.Text, req.Media)
	if err != nil {
		return nil, errs.Wrap(err, "")
	}
	return &proto.WhatsAppMessageResponse{}, nil
}

func (h *whatsApp) QRCode(req *proto.WhatsAppQRRequest, stream proto.WhatsAppService_QRServer) error {
	if req.AccountUUID == "" {
		return errs.New(errors.New(""), errCode.InvalidArgument)
	}
	for {
		if stream.Context().Err() == context.Canceled {
			break
		} else if stream.Context().Err() != nil {
			break
		}
		qr, err := h.service.GetQRCode(req.AccountUUID)
		if err != nil {
			s := status.Convert(err)
			if s.Code() == codes.NotFound {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return errs.Wrap(err, "")
		}
		res := &proto.WhatsAppQRResponse{
			Qr: qr,
		}
		err = stream.Send(res)
		if err != nil {
			return errs.New(err, errCode.Internal)
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}
