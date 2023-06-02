package system

import (
	"context"
	"fmt"
	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"qrpay-wpp/configs"
	"qrpay-wpp/internal/common"
)

type WhatsAppSystem interface {
	Connect(ctx context.Context, accountId string, phone string, eventHandler func(string, any)) error
	GetQRCode(uuid string) (string, error)
	SendMessage(ctx context.Context, uuid string, to string, text string, media []byte) error
}

type whatsAppSystem struct {
	container   *sqlstore.Container
	devices     []*store.Device
	connections Connections
}

func New() (WhatsAppSystem, error) {
	wc := configs.Get().Database
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", wc.Username, wc.Password, wc.Host, wc.Port, wc.Name)
	container, err := sqlstore.New("postgres", url, nil)
	if err != nil {
		return nil, err
	}

	devices, err := container.GetAllDevices()
	if err != nil {
		return nil, err
	}

	return &whatsAppSystem{
		container: container,
		devices:   devices,
	}, nil
}

func (s *whatsAppSystem) getDevice(phone string) *store.Device {
	var device *store.Device
	if phone != "" {
		for _, d := range s.devices {
			if d.ID.User == phone {
				device = d
				break
			}
		}
	}
	if device == nil {
		device = s.container.NewDevice()
	}
	return device
}

func (s *whatsAppSystem) connectNewClient(ctx context.Context, connection *Connection) error {
	client := connection.Client
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return err
	}
	err = client.Connect()
	if err != nil {
		return err
	}
	go s.qrCodeRoutine(ctx, connection, qrChan)
	return nil
}

func (s *whatsAppSystem) Connect(ctx context.Context, accountUUID string, phone string, eventHandler func(string, any)) error {
	existingConnection := s.connections.Get(accountUUID)
	if existingConnection != nil {
		return nil
	}
	device := s.getDevice(phone)
	client := whatsmeow.NewClient(device, nil)
	client.AddEventHandler(func(evt any) {
		eventHandler(accountUUID, evt)
		switch evt.(type) {
		case *events.Disconnected:
			s.restart(accountUUID)
		case *events.TemporaryBan:
			s.restart(accountUUID)
		case *events.LoggedOut:
			s.restart(accountUUID)
		}
	})
	connection := NewConnection(accountUUID, client)
	s.connections.Set(accountUUID, connection)

	if client.Store.ID != nil {
		err := client.Connect()
		if err != nil {
			return err
		}
		return nil
	}
	err := s.connectNewClient(ctx, existingConnection)
	if err != nil {
		return err
	}
	return nil
}

func (s *whatsAppSystem) restart(accountUUID string) {
	connection := s.connections.Get(accountUUID)
	if connection == nil {
		return
	}
	s.connections.Remove(accountUUID)
	s.Connect(context.Background(), accountUUID, "", nil)
}

func (s *whatsAppSystem) qrCodeRoutine(ctx context.Context, connection *Connection, qrChan <-chan whatsmeow.QRChannelItem) {
	var previousCode string
	timeout := false
	for evt := range qrChan {
		if evt.Event != "code" {
			if evt.Event == "timeout" {
				timeout = true
			}
			break
		}
		newQRCode := evt.Code
		if previousCode == newQRCode {
			continue
		}
		connection.QRCode = newQRCode
		previousCode = newQRCode
	}
	if timeout {
		err := s.connectNewClient(ctx, connection)
		if err != nil {
			return
		}
	}
}

func (s *whatsAppSystem) GetQRCode(accountUUID string) (string, error) {
	connection := s.connections.Get(accountUUID)
	if connection == nil {
		return "", status.Error(codes.NotFound, "connection not found")
	}
	return connection.QRCode, nil
}

func (s *whatsAppSystem) SendMessage(ctx context.Context, accountUUID string, phone string, msg string, media []byte) error {
	connection := s.connections.Get(accountUUID)
	if connection == nil {
		return status.Error(codes.NotFound, "connection not found")
	}

	client := connection.Client
	sanitizedPhone := common.SanitizePhone(phone)
	to := types.NewJID(sanitizedPhone, types.DefaultUserServer)

	var message *waProto.Message
	if media != nil {
		res, err := client.Upload(ctx, media, whatsmeow.MediaImage)
		if err != nil {
			return err
		}
		imageMsg := &waProto.ImageMessage{
			Caption:       proto.String(msg),
			Mimetype:      proto.String("image/png"),
			Url:           &res.URL,
			DirectPath:    &res.DirectPath,
			MediaKey:      res.MediaKey,
			FileEncSha256: res.FileEncSHA256,
			FileSha256:    res.FileSHA256,
			FileLength:    &res.FileLength,
		}
		message = &waProto.Message{
			ImageMessage: imageMsg,
		}
	} else {
		message = &waProto.Message{
			Conversation: proto.String(msg),
		}
	}
	_, err := client.SendMessage(ctx, to, message)
	if err != nil {
		return err
	}
	return nil
}
