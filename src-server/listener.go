package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	pb "github.com/modest-as/server-client/grpc"
	sv "github.com/modest-as/server-client/src-server/server"
)

// Listen creates a GRPC server
func Listen(port int, handler sv.Handler) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterCommsServer(s, &commsLayer{handler})

	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}

// commsLayer is used to implement communication GRPC server layer
type commsLayer struct {
	handler sv.Handler
}

// GetNumbers implementation
func (s *commsLayer) GetNumbers(srv pb.Comms_GetNumbersServer) error {
	return s.handler.GetNumbers(srv)
}
