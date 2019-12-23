package main

import (
	"github.com/sitetester/Infocenter/src/stream"
	"net/http"
)

func main() {

	streamer := &stream.Streamer{
		Clients:        make(map[chan stream.Message]bool),
		NewClients:     make(chan (chan stream.Message)),
		DefunctClients: make(chan (chan stream.Message)),
		Messages:       make(chan stream.Message),
	}

	// make stream the HTTP handler for "/infocenter/".  It can do
	// this because it has a ServeHTTP method.  That method
	// is called in a separate goroutine for each
	// request to "/infocenter/".
	http.Handle("/"+stream.ServiceName+"/", streamer)

	// start the server and listen on port 8000
	http.ListenAndServe(":8000", nil)
}
