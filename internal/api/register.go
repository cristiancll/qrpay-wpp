package server

import (
	"google.golang.org/grpc"
	proto "qrpay-wpp/internal/api/proto/generated"
)

func (s *Server) registerServices(grpcServer *grpc.Server) {
	proto.RegisterWhatsAppServiceServer(grpcServer, s.handlers.wpp)
}
