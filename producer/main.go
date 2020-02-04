package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"math/rand"
	"time"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Printf("error making producer: %v", err)
	}
	defer p.Close()
	defer p.Flush(15 * 1000)

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "topic"
	defer p.Flush(15 * 1000)
	for {
		//tightly coupled to consumer and UI, don't do it like this IRL
		sleep := rand.Intn(3)
		
		time.Sleep(time.Duration(sleep) * time.Second)
		id := rand.Intn(7) + 1
		event := rand.Intn(7)
		msg := fmt.Sprintf("%v %v", id, event)
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
		}, nil)
	}

}
