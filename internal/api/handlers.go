package server

import "qrpay-wpp/internal/api/handler"

type handlers struct {
	wpp handler.WhatsApp
}

func (s *Server) createHandlers() {
	s.handlers.wpp = handler.NewWhatsApp(s.services.wpp)
}
