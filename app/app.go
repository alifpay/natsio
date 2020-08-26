package app

import (
	"fmt"

	"github.com/alifpay/natsio/pub"
)

//Server -
type Server struct {
	nm  string
	acc string
	pbc *pub.Client
}

//New -
func New(name, account string, pb *pub.Client) *Server {
	return &Server{nm: name, acc: account, pbc: pb}
}

//ChectAccount -
func (s *Server) ChectAccount(replyTo string, data []byte) {
	fmt.Println(string(data))
	s.pbc.Reply(replyTo, []byte("Account found"))
}
