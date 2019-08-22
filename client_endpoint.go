package mockhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// Client endpoint interface, used by a mock http client for handling outgoing requests
type ClientEndpoint interface {
	http.RoundTripper
	Matches(request *http.Request) bool
}

type clientEndpoint struct {
	requestMatcher requestMatcher
	roundTripFunc  RoundTripFunc
}

// Function for handling a request
type RoundTripFunc func(*http.Request) (*http.Response, error)

// Create a new client endpoint, to be used for configuring a mock http client
func NewClientEndpoint() *clientEndpoint {
	return &clientEndpoint{
		requestMatcher: requestMatcher{},
		roundTripFunc:  responseAsRoundTripFunc(Response()),
	}
}

// Define when this client endpoint should handle a request, according to the provided request matcher
func (e *clientEndpoint) When(matcher *requestMatcher) *clientEndpoint {
	e.requestMatcher = *matcher
	return e
}

// Define the response this client endpoint should return
//
// For more fine grain control, you can use HandleWith function instead.
func (e *clientEndpoint) Respond(response *response) *clientEndpoint {
	return e.HandleWith(responseAsRoundTripFunc(response))
}

// Define an error to return when this client endpoint is triggered. To be used instead of Respond function to mock a
// round trip error (e.g. connection error).
func (e *clientEndpoint) ReturnError(err error) *clientEndpoint {
	return e.HandleWith(func(request *http.Request) (*http.Response, error) {
		return nil, err
	})
}

// Define a round trip function to use for handling a request when this client endpoints is triggered
//
// For simple cases, it is better to simply set a response to return using Respond function, or the ReturnError function
// instead.
func (e *clientEndpoint) HandleWith(roundTrip RoundTripFunc) *clientEndpoint {
	e.roundTripFunc = roundTrip
	return e
}

// Used internally, this is the http.RoundTripper implementation of the client endpoint.
// This is part of the ClientEndpoint interface.
func (e *clientEndpoint) RoundTrip(request *http.Request) (*http.Response, error) {
	return e.roundTripFunc(request)
}

// Used internally to check if this client endpoint matches the given request and should handle it.
// This is part of the ClientEndpoint interface.
func (e *clientEndpoint) Matches(request *http.Request) bool {
	return e.requestMatcher.matches(request)
}

func responseAsRoundTripFunc(r *response) RoundTripFunc {
	return func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: r.statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(r.body)),
			Header:     r.header,
			Request:    request,
		}, nil
	}
}
