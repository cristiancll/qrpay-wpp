package server

import (
	"qrpay-wpp/internal/api/service"
)

type services struct {
	wpp service.WhatsApp
}

func (s *Server) createServices() {
	s.services.wpp = service.NewWhatsApp(s.db, s.repos.wpp)
}
