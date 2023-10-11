package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Broker struct {
	Notifier       chan []byte          // Events are pushed to this channel by the main events-gathering routing
	newClients     chan chan []byte     // new client connections
	closingClients chan chan []byte     // closed client connections
	clients        map[chan []byte]bool // holds currently open connections
}

func (broker *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set headers related to event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each new connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := w.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	// block waiting for messages broadcast on this connection's messageChan
	for {
		// Write to the ResponseWriter
		// Server sent events compatible
		fmt.Fprintf(w, "data: %s\n\n", <-messageChan)

		// Flush the data immediately instead of buffering it for later.
		flusher.Flush()
	}
}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			// A client has connected
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Clients added. %d registered clients", len(broker.clients))

		case s := <-broker.closingClients:
			// A client has deattached and we want to
			// stop sending them messages
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.Notifier:
			// We got a new event from the outside
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				clientMessageChan <- event
			}
		}
	}
}

// Broker Factory
func NewServer() (broker *Broker) {
	// Instantiate the broker
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()
	return broker
}

func main() {
	broker := NewServer()

	go func() {
		for {
			time.Sleep(time.Second * 2)
			eventString := fmt.Sprintf("the time is %v", time.Now())
			log.Println("Receiving event")
			// send data to notifier channel every 2 seconds
			broker.Notifier <- []byte(eventString)
		}
	}()

	log.Fatal("HTTP server error: ", http.ListenAndServe(":3000", broker))
}
