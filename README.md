It's a backend service with go programming language. This service allows clients almost real-time communication by sending messages between each other. Service is able to serve multiple clients concurrently. Each message is sent to a particular topic allowing clients to communicate in separate communication streams.

Messages can be sent by performing HTTP POST request to /infocenter/{topic} route. Topic for the message is passed in the request URL, in place of {topic} tag.

For example client that wants to send a message "labas" to a topic "baras", should perform the following request:

```
POST /infocenter/baras HTTP/1.0
Host: localhost
Content-length: 5

labas
HTTP/1.0 204 No Content
Date: Mon, 14 Sep 2015 08:26:20 GMT


```

Server response is HTTP 204 code if message was accepted successfully. Each message sent to the server is having a unique auto-incrementing ID value.

Message retrieval is available by sending HTTP GET request to the same API route /infocenter/{topic}. Response to this request is an event stream (as defined in W3C specification "Server-Sent Events" http://www.w3.org/TR/eventsource/). All sent messages  have a message type "msg". Service disconnect all the clients if they were consuming the stream for more than max allowed time e.g. 30 sec. Before client is disconnected, server send special "timeout" event. The contents of the "timeout" event is the time how long client was connected.

Example of receiving message events:
```
GET /infocenter/baras HTTP/1.0
Host: localhost

HTTP/1.0 200 OK
Cache-Control: no-cache
Content-Type: text/event-stream
Date: Mon, 14 Sep 2015 08:33:46 GMT

id: 7
event: msg
data: labas

event: timeout
data: 30s


```

Clients do not need to create or destroy topics before starting to use them. Just by getting or writing messages to a particular topic automatically create a communication channel. When clients not using a topic, all resources allocated are freed.
