package mockhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type ClientEndpoint interface {
	http.RoundTripper
	Matches(request *http.Request) bool
}

type clientEndpoint struct {
	requestMatcher requestMatcher
	roundTripFunc  RoundTripFunc
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func NewClientEndpoint() *clientEndpoint {
	return &clientEndpoint{
		requestMatcher: requestMatcher{},
		roundTripFunc:  responseAsRoundTripFunc(Response()),
	}
}

func (e *clientEndpoint) When(matcher *requestMatcher) *clientEndpoint {
	e.requestMatcher = *matcher
	return e
}

func (e *clientEndpoint) Respond(response *response) *clientEndpoint {
	return e.HandleWith(responseAsRoundTripFunc(response))
}

func (e *clientEndpoint) ReturnError(err error) *clientEndpoint {
	return e.HandleWith(func(request *http.Request) (*http.Response, error) {
		return nil, err
	})
}

func (e *clientEndpoint) HandleWith(roundTrip RoundTripFunc) *clientEndpoint {
	e.roundTripFunc = roundTrip
	return e
}

func (e *clientEndpoint) RoundTrip(request *http.Request) (*http.Response, error) {
	return e.roundTripFunc(request)
}

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
