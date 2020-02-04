package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	
	"log"
	"net/http"
)

// Example SSE server in Golang.
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte
	// New client connections
	newClients chan chan []byte
	// Closed client connections
	closingClients chan chan []byte
	// Client connections registry
	clients map[chan []byte]bool
}

func NewServer() (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := req.Context().Done()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()
	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		fmt.Fprintf(rw, "data: %v\n\n", string(<-messageChan))

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan, _ := range broker.clients {
				clientMessageChan <- event
			}
		}
	}

}

func main() {
	broker := NewServer()
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/events", broker)
	r.Handle("/", http.FileServer(http.Dir("./client")))
	
	//kafka setup
	
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id": "somegroup",
	})
	if err != nil {
		log.Fatalf("error setting up consumer: %v", err)
	}
	
	err = consumer.Subscribe("topic", nil)
	if err != nil {
		log.Printf("err subscribing: %v", err)
	}
	
	go func() {
		fmt.Println("running")
		for {
			msg, err := consumer.ReadMessage(-1) //looking only at messages because it's a demo
			if err != nil {
				log.Printf("error reading msg: %v", err)
			}
			broker.Notifier <- msg.Value
		}
	}()
	log.Fatal("HTTP server error: ", http.ListenAndServe(":3000", r))
}
