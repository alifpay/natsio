package sub

import (
	"github.com/nats-io/nats.go"
)

func (c *Client) checkGOB(m *nats.Msg) {
	go c.srv.Check(m.Reply, "gob", m.Data)
}

func (c *Client) checkJSON(m *nats.Msg) {
	go c.srv.Check(m.Reply, "json", m.Data)
}
