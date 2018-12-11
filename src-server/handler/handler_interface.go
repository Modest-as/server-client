package handler

import pb "github.com/modest-as/server-client/grpc"

// Handler is a shared interface for stateful/less implementations
type Handler interface {
	GetNumbers(srv pb.Comms_GetNumbersServer) error
}
