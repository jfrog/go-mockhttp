package mockhttp

import (
	"net/http"
	"time"
)

type ServerEndpoint interface {
	http.Handler
	Matches(request *http.Request) bool
}

type serverEndpoint struct {
	requestMatcher requestMatcher
	handlerFunc    http.HandlerFunc
}

func NewServerEndpoint() *serverEndpoint {
	return &serverEndpoint{
		requestMatcher: requestMatcher{},
		handlerFunc:    responseAsHandler(Response()),
	}
}

func (e *serverEndpoint) When(matcher *requestMatcher) *serverEndpoint {
	e.requestMatcher = *matcher
	return e
}

func (e *serverEndpoint) Respond(response *response) *serverEndpoint {
	return e.HandleWith(responseAsHandler(response))
}

func (e *serverEndpoint) HandleWith(handlerFunc http.HandlerFunc) *serverEndpoint {
	e.handlerFunc = handlerFunc
	return e
}

func (e *serverEndpoint) Matches(request *http.Request) bool {
	return e.requestMatcher.matches(request)
}

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
