package pub

import (
	"context"
	"crypto/tls"
	"log"
	"strings"
	"sync"
	"time"

	"git.alifpay.tj/providers/core/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

//Client of nats streaming server
type Client struct {
	nc   *nats.Conn
	lock sync.RWMutex
	sc   stan.Conn
	cfg  models.Config
}

//New - init new nats client for publish
func New(conf models.Config) *Client {
	return &Client{cfg: conf}
}

//Run connects to nats server
func (c *Client) Run(ctx context.Context, tlsCfg *tls.Config) {
	var err error

	// Setup the connect options
	opts := []nats.Option{nats.Name("providers-pub")}
	// Use UserCredentials
	//if *userCreds != "" {
	//opts = append(opts, nats.UserCredentials(*userCreds))
	//}
	if strings.Contains(c.cfg.NatsURL, "tls://") {
		opts = append(opts, nats.Secure(tlsCfg))
	}
	//nats.ErrorHandler(logSlowConsumer)
	opts = append(opts, nats.ReconnectWait(5*time.Second))
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Println("pub DisconnectErrHandler", err)
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("pub Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Println("pub ClosedHandler", nc.LastError())
	}))

	c.nc, err = nats.Connect(c.cfg.NatsURL, opts...)
	if err != nil {
		log.Fatalln("pub nats.Connect", c.cfg.NatsURL, err)
		return
	}
	log.Println("pub nats client is connected")
	c.streaming()
	<-ctx.Done()
	c.close()
}

//Close nats client
func (c *Client) close() {
	if c.nc != nil {
		c.sc.Close()
		c.nc.Drain()
	}
}

//streaming -
func (c *Client) streaming() {
	var err error
	c.lock.Lock()
	c.sc, err = stan.Connect(c.cfg.ClusterID, c.cfg.ClientID+"-pub", stan.NatsConn(c.nc),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v", reason)
			time.Sleep(15 * time.Second)
			c.sc.Close()
			c.streaming()
		}))
	c.lock.Unlock()
	if err != nil {
		log.Printf("stan.Connect: %v", err)
	}
}

//Reply - publish response of request
func (c *Client) Reply(replyTo string, data []byte) {
	err := c.nc.Publish(replyTo, data)
	c.nc.Flush()
	if err != nil {
		log.Println("nc.Publish", err)
	}
}

//PublishStream - publish data to nats streaming server
func (c *Client) PublishStream(subj string, data []byte) {
	err := c.sc.Publish(subj, data)
	if err != nil {
		log.Println("sc.Publish", err)
	}
}
