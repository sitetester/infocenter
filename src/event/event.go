// Event related stuff will go here.
package event

import (
	"fmt"
)

type StandardEvent struct {
	Id    int
	Event string
	Data  string
}

type Formatter interface {
	// since SSE supports JSON & XML formats as well
	// this is to make our code more flexible, so we could format it in multiple ways (e.g. standard plain text, json or xml)
	Format(event StandardEvent) string
}

type StandardFormatter struct{}

func (standardFormatter StandardFormatter) Format(event StandardEvent) string {
	return fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", event.Id, event.Event, event.Data)
}

// TODO: In similar way, we could have our JsonFormatter & XmlFormatter

type TimedOutEvent struct {
	Event string
	Data  string
}

type TimedOutFormatter interface {
	Format(event TimedOutEvent) string
}

type TimedOutEventFormatter struct{}

func (timedOutEventFormatter TimedOutEventFormatter) Format(event TimedOutEvent) string {
	return fmt.Sprintf("event: %s\ndata %s\n\n", event.Event, event.Data)
}

// TODO: In similar way, we could have our TimedOutJsonFormatter & TimedOutXmlFormatter
