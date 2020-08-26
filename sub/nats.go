package sub

import (
	"context"
	"crypto/tls"
	"log"
	"strings"
	"sync"
	"time"

	"git.alifpay.tj/providers/core/app"
	"git.alifpay.tj/providers/core/models"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

//Client of nats streaming server
type Client struct {
	nc   *nats.Conn
	lock sync.RWMutex
	sc   stan.Conn
	srv  *app.Server
	cfg  models.Config
}

//New - init new nats client for subscribe
func New(s *app.Server, conf models.Config) *Client {
	return &Client{srv: s, cfg: conf}
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
	if strings.Contains(c.cfg.NatsURL, "tls://") {
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

	c.nc, err = nats.Connect(c.cfg.NatsURL, opts...)
	if err != nil {
		log.Fatalln("sub nats.Connect", c.cfg.NatsURL, err)
		return
	}

	//Subscribe check account Request-Reply
	_, err = c.nc.Subscribe(c.cfg.CheckSubject+"-gob", c.checkGOB)
	if err != nil {
		log.Fatalln("nc.Subscribe c.CheckSubject gob", c.cfg.CheckSubject, err)
		return
	}

	//Subscribe check account Request-Reply json encoding
	_, err = c.nc.Subscribe(c.cfg.CheckSubject+"-json", c.checkJSON)
	if err != nil {
		log.Fatalln("nc.Subscribe c.CheckSubject json", c.cfg.CheckSubject, err)
		return
	}

	//Subscribe status of payment Request-Reply
	_, err = c.nc.Subscribe(c.cfg.StatusSubject+"-gob", c.statusGOB)
	if err != nil {
		log.Fatalln("nc.Subscribe StatusSubject gob", c.cfg.StatusSubject, err)
		return
	}

	//Subscribe status of payment Request-Reply json encoding
	_, err = c.nc.Subscribe(c.cfg.StatusSubject+"-json", c.statusJSON)
	if err != nil {
		log.Fatalln("nc.Subscribe StatusSubject json", c.cfg.StatusSubject, err)
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
	c.sc, err = stan.Connect(c.cfg.ClusterID, c.cfg.ClientID+"-sub", stan.NatsConn(c.nc),
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
	_, err = c.sc.Subscribe(c.cfg.PaySubject+"-gob", c.paymentsGOB, stan.DurableName(c.cfg.DurableName))
	if err != nil {
		log.Printf("sc.Subscribe: %v", err)
	}

	// Subscribe payment request subject json encoding with durable name
	_, err = c.sc.Subscribe(c.cfg.PaySubject+"-json", c.paymentsJSON, stan.DurableName(c.cfg.DurableName))
	if err != nil {
		log.Printf("sc.Subscribe: %v", err)
	}

	// Subscribe payment retry subject with durable name
	_, err = c.sc.Subscribe(c.cfg.RetrySubject, c.paymentsJSON, stan.DurableName(c.cfg.DurableName))
	if err != nil {
		log.Printf("sc.Subscribe: %v", err)
	}

	// Subscribe payment status check subject with durable name
	_, err = c.sc.Subscribe(c.cfg.StatusCheckSubject, c.paymentsJSON, stan.DurableName(c.cfg.DurableName))
	if err != nil {
		log.Printf("sc.Subscribe: %v", err)
	}
}
