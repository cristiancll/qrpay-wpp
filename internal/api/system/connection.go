package system

import (
	"go.mau.fi/whatsmeow"
)

type Connections map[string]*Connection

func (c Connections) Get(accountId string) *Connection {
	return c[accountId]
}
func (c Connections) Set(accountId string, connection *Connection) {
	if accountId != "" && connection != nil {
		c[accountId] = connection
	}
}

type Connection struct {
	AccountId string
	Client    *whatsmeow.Client
	QRCode    string
	Connected bool
	Paired    bool
}

func NewConnection(accountId string, client *whatsmeow.Client) *Connection {
	return &Connection{
		AccountId: accountId,
		Client:    client,
	}
}
