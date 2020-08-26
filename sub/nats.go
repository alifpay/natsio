package sub

import (
	"context"
	"crypto/tls"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/alifpay/natsio/app"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

//Client of nats streaming server
type Client struct {
	nc           *nats.Conn
	lock         sync.RWMutex
	sc           stan.Conn
	srv          *app.Server
	natsURL      string
	checkSubject string
	clusterID    string
	clientID     string
	paySubject   string
	durableName  string
}

//New - init new nats client for subscribe
func New(s *app.Server, natsURL, checkSubject, cluster, clientid, paySubject, durableName string) *Client {
	return &Client{srv: s, natsURL: natsURL, checkSubject: checkSubject, clusterID: cluster, paySubject: paySubject, durableName: durableName}
}

//Run connects to nats server
func (c *Client) Run(ctx context.Context, tlsCfg *tls.Config) {
	var err error

	// Setup the connect options
	opts := []nats.Option{nats.Name("providers-sub")}
	// Use UserCredentials
	//if *userCreds != "" {
	//opts = append(opts, nats.UserCredentials(*userCreds))
	//}
	if strings.Contains(c.natsURL, "tls://") {
		opts = append(opts, nats.Secure(tlsCfg))
	}
	//nats.ErrorHandler(logSlowConsumer)
	opts = append(opts, nats.ReconnectWait(5*time.Second))
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Println("sub disconnectErrHandler", err)
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("sub reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Println("sub closedHandler", nc.LastError())
	}))

	c.nc, err = nats.Connect(c.natsURL, opts...)
	if err != nil {
		log.Fatalln("sub nats.Connect", c.natsURL, err)
		return
	}

	//Subscribe check account Request-Reply
	_, err = c.nc.Subscribe(c.checkSubject, c.check)
	if err != nil {
		log.Fatalln("nc.Subscribe c.CheckSubject gob", c.checkSubject, err)
		return
	}

	log.Println("sub nats client is connected")
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
	c.sc, err = stan.Connect(c.clusterID, c.clientID+"-sub", stan.NatsConn(c.nc),
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

	// Subscribe payment request subject gob encoding with durable name
	_, err = c.sc.Subscribe(c.paySubject, c.payment, stan.DurableName(c.durableName))
	if err != nil {
		log.Printf("sc.Subscribe: %v", err)
	}

}
