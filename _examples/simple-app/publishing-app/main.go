package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/infrastructure/kafka"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/renstrom/shortuuid"
)

var (
	brokers = []string{"kafka:9092"}
)

func main() {
	log.Println("Starting publishing app")

	publisher, err := kafka.NewPublisher(brokers, kafka.DefaultMarshaler{}, nil)
	if err != nil {
		panic(err)
	}
	defer publisher.Close()

	messagesToAdd := 10000
	workers := 25

	msgAdded := make(chan struct{})
	allMessagesAdded := make(chan struct{})

	go func() {
		for range msgAdded {
			messagesToAdd--

			if messagesToAdd%1000 == 0 {
				log.Println("left ", messagesToAdd)
			}
			if messagesToAdd == 0 {
				allMessagesAdded <- struct{}{}
			}
		}
	}()

	for num := 0; num < workers; num++ {
		go func() {
			var msgPayload postAdded
			var msg *message.Message

			for messagesToAdd > 0 {
				msgPayload.OccurredOn = time.Now()
				msgPayload.Author = randString(10)
				msgPayload.Title = randString(15)
				msgPayload.Content = randString(30)

				payload, err := json.Marshal(msgPayload)
				if err != nil {
					panic(err)
				}

				msg = message.NewMessage(uuid.NewV4().String(), payload)

				// using function from middleware to set correlation id, useful for debugging
				middleware.SetCorrelationID(shortuuid.New(), msg)

				err = publisher.Publish("posts_published", msg)
				if err != nil {
					log.Println("cannot publish message:", err)
					continue
				}
				msgAdded <- struct{}{}
			}
		}()
	}

	// waiting to all being produced
	<-allMessagesAdded
}

type postAdded struct {
	OccurredOn time.Time `json:"occurred_on"`

	Author string `json:"author"`
	Title  string `json:"title"`

	Content string `json:"content"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

// randString generates random string of len n
func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
