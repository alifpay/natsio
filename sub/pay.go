package sub

import "github.com/nats-io/stan.go"

func (c *Client) paymentsGOB(m *stan.Msg) {
	c.srv.Pay("gob", m.Data)
}

func (c *Client) paymentsJSON(m *stan.Msg) {
	c.srv.Pay("json", m.Data)
}

func (c *Client) paymentRetry(m *stan.Msg) {
	c.srv.Pay("json", m.Data)
}
