package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alifpay/natsio/app"
	"github.com/alifpay/natsio/pub"
	"github.com/alifpay/natsio/sub"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	//publisher nats client
	np := pub.New("nats://127.0.0.1:4222", "alif-dev", "dev-test")
	go np.Run(ctx, nil)

	srv := app.New("Test", "123456789", np)

	//subscriber nats client
	ns := sub.New(srv, "nats://127.0.0.1:4222", "test-check-acc", "alif-dev", "dev-test", "test-pay", "testclient0")
	go ns.Run(ctx, nil)
	log.Println("server is running")
	//catches signal OS interruption and sends cancel to context
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)

	<-sigs
	log.Println("shutdown signal received")
	signal.Stop(sigs)
	close(sigs)
	cancel()

}
