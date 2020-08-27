package app

import (
	"github.com/alifpay/natsio/feedpb"
	"github.com/alifpay/natsio/pub"
)

//Server -
type Server struct {
	str feedpb.Feeds_BroadcastServer
	pbc *pub.Client
}

//New -
func New(pb *pub.Client) *Server {
	return &Server{pbc: pb}
}
