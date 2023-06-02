package system

import (
	"go.mau.fi/whatsmeow"
)

type Connections map[string]*Connection

func (c Connections) Get(accountUUID string) *Connection {
	return c[accountUUID]
}
func (c Connections) Set(accountUUID string, connection *Connection) {
	if accountUUID != "" && connection != nil {
		c[accountUUID] = connection
	}
}

func (c Connections) Remove(accountUUID string) {
	delete(c, accountUUID)
}

type Connection struct {
	AccountUUID string
	Client      *whatsmeow.Client
	QRCode      string
	Connected   bool
	Paired      bool
}

func NewConnection(accountUUID string, client *whatsmeow.Client) *Connection {
	return &Connection{
		AccountUUID: accountUUID,
		Client:      client,
	}
}
