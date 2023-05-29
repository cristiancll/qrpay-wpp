package server

import (
	"fmt"
	"qrpay-wpp/internal/api/repository"
)

type repositories struct {
	wpp repository.WhatsApp
}

func (s *Server) createRepositories() error {
	s.repos.wpp = repository.NewWhatsApp(s.db)
	if err := s.repos.wpp.Migrate(s.context); err != nil {
		return fmt.Errorf("unable to migrate whatsapp repository: %v", err)
	}
	return nil
}
