package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	proto "qrpay-wpp/internal/api/proto/generated"
	"qrpay-wpp/internal/api/service"
	"qrpay-wpp/internal/errors"
	"time"
)

type WhatsApp interface {
	Get(ctx context.Context, req *proto.WhatsAppGetRequest) (*proto.WhatsAppGetResponse, error)
	List(ctx context.Context, req *proto.VoidRequest) (*proto.WhatsAppListResponse, error)
	QRCodeStream(req *proto.VoidRequest, stream proto.WhatsAppService_QRCodeStreamServer) error
	proto.WhatsAppServiceServer
}

type whatsApp struct {
	service service.WhatsApp
	proto.UnimplementedWhatsAppServiceServer
}

func NewWhatsApp(s service.WhatsApp) WhatsApp {
	return &whatsApp{service: s}
}

func (h *whatsApp) Get(ctx context.Context, req *proto.WhatsAppGetRequest) (*proto.WhatsAppGetResponse, error) {
	if req.Uuid == "" {
		return nil, status.Error(codes.InvalidArgument, errors.UUID_REQUIRED)
	}
	whats, err := h.service.Get(ctx, req.Uuid)
	if err != nil {
		return nil, err
	}
	res := &proto.WhatsAppGetResponse{
		WhatsApp: &proto.WhatsApp{
			Uuid:      whats.UUID,
			Qr:        whats.QR,
			Phone:     whats.Phone,
			Active:    whats.Active,
			Banned:    whats.Banned,
			CreatedAt: timestamppb.New(whats.CreatedAt),
			UpdatedAt: timestamppb.New(whats.UpdatedAt),
		},
	}
	return res, nil
}

func (h *whatsApp) List(ctx context.Context, _ *proto.VoidRequest) (*proto.WhatsAppListResponse, error) {
	whatsList, err := h.service.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	pWhatsList := make([]*proto.WhatsApp, 0)
	for _, whats := range whatsList {
		pWhatsList = append(pWhatsList, &proto.WhatsApp{
			Uuid:      whats.UUID,
			Qr:        whats.QR,
			Phone:     whats.Phone,
			Scanned:   whats.Scanned,
			Active:    whats.Active,
			Banned:    whats.Banned,
			CreatedAt: timestamppb.New(whats.CreatedAt),
			UpdatedAt: timestamppb.New(whats.UpdatedAt),
		})
	}
	res := &proto.WhatsAppListResponse{
		WhatsAppList: pWhatsList,
	}
	return res, nil
}

func (h *whatsApp) QRCodeStream(_ *proto.VoidRequest, stream proto.WhatsAppService_QRCodeStreamServer) error {
	wpp, err := h.service.GetActiveWhatsApp(stream.Context())
	if err == nil {
		res := &proto.WhatsApp{
			Uuid:      wpp.UUID,
			Qr:        wpp.QR,
			Phone:     wpp.Phone,
			Scanned:   wpp.Scanned,
			Active:    wpp.Active,
			Banned:    wpp.Banned,
			CreatedAt: timestamppb.New(wpp.CreatedAt),
			UpdatedAt: timestamppb.New(wpp.UpdatedAt),
		}
		err = stream.Send(res)
		if err != nil {
			return err
		}
		return nil
	}
	for {
		if stream.Context().Err() == context.Canceled {
			break
		} else if stream.Context().Err() != nil {
			break
		}
		wpp, err = h.service.GetUnscannedWhatsApp(stream.Context())
		if err != nil {
			s := status.Convert(err)
			if s.Code() == codes.NotFound {
				time.Sleep(5 * time.Second)
				continue
			}
			return err
		}
		res := &proto.WhatsApp{
			Uuid:      wpp.UUID,
			Qr:        wpp.QR,
			Phone:     wpp.Phone,
			Scanned:   wpp.Scanned,
			Active:    wpp.Active,
			Banned:    wpp.Banned,
			CreatedAt: timestamppb.New(wpp.CreatedAt),
			UpdatedAt: timestamppb.New(wpp.UpdatedAt),
		}
		err = stream.Send(res)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}
