package mockhttp

import (
	"net/http"
	"time"
)

// Server endpoint interface, used by a mock http server for handling incoming requests
type ServerEndpoint interface {
	http.Handler
	Matches(request *http.Request) bool
}

type serverEndpoint struct {
	requestMatcher requestMatcher
	handlerFunc    http.HandlerFunc
}

// Create a new server endpoint, to be used for configuring a mock http server
func NewServerEndpoint() *serverEndpoint {
	return &serverEndpoint{
		requestMatcher: requestMatcher{},
		handlerFunc:    responseAsHandler(Response()),
	}
}

// Define when this server endpoint should handle a request, according to the provided request matcher
func (e *serverEndpoint) When(matcher *requestMatcher) *serverEndpoint {
	e.requestMatcher = *matcher
	return e
}

// Define the response this server endpoint should return
//
// For more fine grain control, you can use HandleWith function instead.
func (e *serverEndpoint) Respond(response *response) *serverEndpoint {
	return e.HandleWith(responseAsHandler(response))
}

// Define a http handler function to use for sending a response when this server endpoints is triggered
//
// For simple cases, it is better to simply set a response to send using Respond function instead.
func (e *serverEndpoint) HandleWith(handlerFunc http.HandlerFunc) *serverEndpoint {
	e.handlerFunc = handlerFunc
	return e
}

// Used internally to check if this server endpoint matches the given request and should handle it.
// This is part of the ServerEndpoint interface.
func (e *serverEndpoint) Matches(request *http.Request) bool {
	return e.requestMatcher.matches(request)
}

// Used internally, this is the http.Handler implementation of the server endpoint.
// This is part of the ServerEndpoint interface.
func (e *serverEndpoint) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	e.handlerFunc(response, request)
}

func responseAsHandler(r *response) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if r.delay > 0 {
			time.Sleep(r.delay)
		}
		for name, values := range r.header {
			for _, v := range values {
				response.Header().Add(name, v)
			}
		}
		response.WriteHeader(r.statusCode)
		response.Write(r.body)
	}
}
