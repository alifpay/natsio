package sub

import (
	"github.com/nats-io/nats.go"
)

func (c *Client) check(m *nats.Msg) {
	go c.srv.GetFeed(m.Reply, m.Data)
}
