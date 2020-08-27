package app

import (
	"fmt"
	"io"
	"log"

	"github.com/alifpay/natsio/feedpb"
)

//Broadcast -
func (s *Server) Broadcast(stream feedpb.Feeds_BroadcastServer) error {
	s.str = stream
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Println("could not recieve from stream : ", err)
			return err
		}

		feed := "New Feed Recieved: " + msg.GetFeed()
		s.pbc.Reply(msg.ReplyTo, []byte(feed))
	}
}

//GetFeed -
func (s *Server) GetFeed(to string, msg []byte) {
	fmt.Println("sending new feed...")
	err := s.str.Send(&feedpb.FeedResponse{Feed: string(msg), ReplyTo: to})
	if err != nil {
		log.Println("s.str.Send", err)
	}
}
