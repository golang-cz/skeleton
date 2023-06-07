package scheduler

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"time"
)

func Run() {

	// nats-server url as environment variable
	nurl := os.Getenv("NATS_URL")

	// Connect to NATS server
	nc, err := nats.Connect(nurl)

	if err != nil {
		log.Fatal(fmt.Errorf("failed to create a nats-server: %w", err))
	}

	log.Printf("created connection with nats-server")
	defer nc.Close()

	// Create new inbox
	inbox := nats.NewInbox()
	sub, err := nc.SubscribeSync(inbox)

	if err != nil {
		log.Fatal(fmt.Errorf("failed to create a new inbox: %w", err))
	}
	log.Printf("created new inbox")

	// Unsubscription function
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			log.Fatal(fmt.Errorf("failed to unsubscribe from inbox: %w", err))
		}
	}()

	// Subscribe to "ping" topic and automatically reply with pong message
	_, err = nc.Subscribe("ping", func(m *nats.Msg) {
		if err := m.Respond([]byte("pong")); err != nil {
			log.Fatal(fmt.Errorf("failed to send the pong message: %w", err))
		}
	})

	if err != nil {
		log.Fatal(fmt.Errorf("failed to create a subscibtion to the ping requests: %w", err))
	}

	// Send ping message to all microservices
	err = nc.PublishRequest("ping", inbox, []byte("ping"))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to send a ping request: %w", err))
	}
	log.Printf("send a ping request")

	cnt := 0
	for {
		msg, _ := sub.NextMsg(time.Second * 2)
		if msg == nil {
			break
		}
		cnt++
	}

	// Print number of pongs received
	fmt.Printf("Number of pongs received: %d\n", cnt)

	select {}

}
