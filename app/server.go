package app

import (
	"net"

	"github.com/alifpay/natsio/feedpb"
	"google.golang.org/grpc"
)

var defaultServer *grpc.Server

//Run serve secure grpc Server
func (s *Server) Run(host string) error {

	// Create the channel to listen on
	l, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	defaultServer = grpc.NewServer()

	feedpb.RegisterFeedsServer(defaultServer, s)

	err = defaultServer.Serve(l)
	if err != nil {
		return err
	}

	return nil
}

//Close stop grpc Server
func Close() {
	defaultServer.GracefulStop()
}
