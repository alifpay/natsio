package sub

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

//check status of payments
func (c *Client) statusGOB(m *nats.Msg) {

}

//check status of payments
func (c *Client) statusJSON(m *nats.Msg) {

}

func (c *Client) paymentStatusCheck(m *stan.Msg) {
	c.srv.Status(m.Data)
}
