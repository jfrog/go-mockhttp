package mockhttp

import (
	"net/http"
	"time"
)

type response struct {
	statusCode int
	body       []byte
	header     http.Header
	delay      time.Duration
}

// Response creates a new response definition.
// To be used e.g. with the ServerEndpoint's Respond builder function.
//
// Defaults:
//   - Status code: OK 200
//   - Empty body
//   - No headers
//   - No delay
func Response() *response {
	return &response{
		statusCode: http.StatusOK,
		body:       []byte{},
		header:     http.Header{},
	}
}

// StatusCode sets the status code to respond with
func (r *response) StatusCode(statusCode int) *response {
	r.statusCode = statusCode
	return r
}

// Body sets the body bytes to respond with
func (r *response) Body(body []byte) *response {
	r.body = body
	return r
}

// BodyString sets the body (as string) to respond with
func (r *response) BodyString(body string) *response {
	r.body = []byte(body)
	return r
}

// Header sets header key-values pair to respond with
func (r *response) Header(key, value string, other ...string) *response {
	r.header.Set(key, value)
	for _, v := range other {
		r.header.Add(key, v)
	}
	return r
}

// Delay sets a delay, after receiving a request, before sending the response
func (r *response) Delay(delay time.Duration) *response {
	r.delay = delay
	return r
}
