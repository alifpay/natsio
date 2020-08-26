package sub

import (
	"github.com/nats-io/nats.go"
)

func (c *Client) check(m *nats.Msg) {
	go c.srv.ChectAccount(m.Reply, m.Data)
	//m.Respond("Test account found")
}
