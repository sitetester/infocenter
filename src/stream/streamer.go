// This will manage broadcasting POSTed message to connected clients.
package stream

import (
	"fmt"
	"github.com/sitetester/Infocenter/src/event"
	"github.com/sitetester/Infocenter/src/helper"
	"log"
	"net/http"
	"time"
)

// Stores HTTP GET request connection time & subscribed topic
type SubscriptionInfo struct {
	ConnectedAt time.Time
	Topic       string
}

type Message struct {
	Id    int
	Topic string
	Msg   string
}

// ```All sent Messages should have a message type "msg".```
const EventType = "msg"

// ```Before client is disconnected, server should send special "timeout" event.```
const TimedOutEventType = "timeout"

// Service should disconnect all the clients if they were consuming the stream for more than max allowed time (e.g. 30 sec).
const MaxAllowedConnectionTime = 30

// ```Lets call this service Infocenter.```
const ServiceName = "infocenter"

// a single Streamer will be created in this program. It is responsible
// for keeping a list of which Clients (browsers) are currently attached
// and broadcasting events (Messages) to those Clients.

type Streamer struct {

	// create a map of Clients, the keys of the map are the channels
	// over which we can push messages to attached clients. (The values
	// are just booleans and are meaningless)
	Clients map[chan Message]bool

	// channel into which new clients can be pushed
	NewClients chan chan Message

	// channel into which disconnected clients should be pushed
	DefunctClients chan chan Message

	// channel into which messages are pushed to be broadcast out to attached clients
	Messages chan Message

	// this will keep track of auto incrementing each POSTed message ID
	MsgId int
}

// handles the addition & removal of Clients, as well as the forwarding of messages to attached clients
func (streamer *Streamer) start() {

	go func() {

		for {
			// block until we receive from one of the three following channels
			select {

			case messageChan := <-streamer.NewClients:

				// there is a new client attached and we
				// want to start sending them messages
				streamer.Clients[messageChan] = true
				log.Println("Added new client")

			case messageChan := <-streamer.DefunctClients:
				// a client has detached and we want to stop sending them messages
				delete(streamer.Clients, messageChan)
				log.Println("Removed client")

				_, isOpen := <-messageChan
				if isOpen {
					close(messageChan)
				}

			case message := <-streamer.Messages:
				// there is a new Message to send. For each
				// attached client, push the new Message
				// into the client's clientMessageChan channel
				for clientMessageChan := range streamer.Clients {
					clientMessageChan <- message
				}
			}
		}
	}()
}

func (streamer *Streamer) publishMsg(responseWriter http.ResponseWriter, request *http.Request, topic string) {

	msg := request.FormValue("msg")
	log.Printf("Message received for topic[%s]: %s", topic, msg)

	// ```Server response should be HTTP 204 code if message was accepted successfully.```
	responseWriter.WriteHeader(http.StatusNoContent)

	// ```Each message sent to the server should have unique auto-incrementing ID value.```
	streamer.MsgId += 1

	postedMsg := Message{
		Id:    streamer.MsgId,
		Topic: topic,
		Msg:   msg,
	}

	go func() {
		streamer.Messages <- postedMsg
	}()
}

func (streamer *Streamer) formatEvent(formatter event.Formatter, event event.StandardEvent) string {
	return formatter.Format(event)
}

func (streamer *Streamer) formatTimedOutEvent(timedOutFormatter event.TimedOutFormatter, event event.TimedOutEvent) string {
	return timedOutFormatter.Format(event)
}

func (streamer *Streamer) streamTopicMsg(
	responseWriter http.ResponseWriter,
	messageChan chan Message,
	flusher http.Flusher,
	subscriptionInfo SubscriptionInfo) {

	for {

		select {

		case message, _ := <-messageChan:
			if isTimedOut(subscriptionInfo) {
				streamer.handleTimeout(responseWriter, flusher)

				go func() {
					streamer.DefunctClients <- messageChan
				}()

				// https://stackoverflow.com/questions/25469682/break-out-of-select-loop
				return
			}

			if message.Topic == subscriptionInfo.Topic {
				standardEvent := event.StandardEvent{
					Id:    message.Id,
					Event: EventType,
					Data:  message.Msg,
				}

				var standardFormatter event.StandardFormatter
				_, err := fmt.Fprintf(responseWriter, streamer.formatEvent(standardFormatter, standardEvent))
				if err != nil {
					log.Fatal(err)
				}

				flusher.Flush()
			}

		default:
			if isTimedOut(subscriptionInfo) {
				streamer.handleTimeout(responseWriter, flusher)

				go func() {
					streamer.DefunctClients <- messageChan
				}()

				// https://stackoverflow.com/questions/25469682/break-out-of-select-loop
				return
			}
		}
	}
}

// ```Service should disconnect all the clients if they were consuming the stream for more than max allowed time(e.g. 30 sec).```
func isTimedOut(subscriptionInfo SubscriptionInfo) bool {

	diffInt := (int)(time.Now().Sub(subscriptionInfo.ConnectedAt).Seconds())
	return diffInt > MaxAllowedConnectionTime
}

// ```Before client is disconnected, server should send special "timeout" event.```
// ```The contents of the "timeout" event should be the time how long client was connected.```
func (streamer *Streamer) handleTimeout(responseWriter http.ResponseWriter, flusher http.Flusher) {

	timedOutEvent := event.TimedOutEvent{
		Event: TimedOutEventType,
		Data:  fmt.Sprintf("%ds", MaxAllowedConnectionTime),
	}

	var timedOutEventFormatter event.TimedOutEventFormatter
	streamer.formatTimedOutEvent(timedOutEventFormatter, timedOutEvent)

	_, err := fmt.Fprintf(responseWriter, timedOutEventFormatter.Format(timedOutEvent))
	if err != nil {
		log.Fatal(err)
	}

	flusher.Flush()
}

// handles HTTP request at the "/infocenter/"
func (streamer *Streamer) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	streamer.start()

	var regexHelper helper.Regex
	topic := regexHelper.ParseTopic(request.RequestURI, ServiceName)
	if topic == "" {
		fmt.Println("Error parsing topic from request URI.")
		return
	}

	// make sure the writer supports flushing
	flusher, ok := responseWriter.(http.Flusher)
	if !ok {
		http.Error(responseWriter, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	if request.Method == "GET" {
		// set essential HTTP streaming related headers
		setHeaders(responseWriter)

		// create a new channel, over which the streamer can send this client messages
		messageChan := make(chan Message)

		// add this client to the map of those that should receive updates
		streamer.NewClients <- messageChan

		subscriptionInfo := SubscriptionInfo{
			ConnectedAt: time.Now(),
			Topic:       topic,
		}

		streamer.streamTopicMsg(responseWriter, messageChan, flusher, subscriptionInfo)
	}

	if request.Method == "POST" {
		streamer.publishMsg(responseWriter, request, topic)
	}
}

func setHeaders(responseWriter http.ResponseWriter) {

	responseWriter.Header().Set("Content-Type", "text/event-stream")
	responseWriter.Header().Set("Cache-Control", "no-cache")
	responseWriter.Header().Set("Connection", "keep-alive")
	responseWriter.Header().Set("Transfer-Encoding", "chunked")
}
