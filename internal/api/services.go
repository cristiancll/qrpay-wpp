package server

import (
	"qrpay-wpp/internal/api/service"
	"qrpay-wpp/internal/api/system"
)

type services struct {
	wpp service.WhatsApp
}

func (s *Server) createServices(wppSystem system.WhatsAppSystem) {
	s.services.wpp = service.NewWhatsApp(s.db, s.repos.wpp, wppSystem)
}
