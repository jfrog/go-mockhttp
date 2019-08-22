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

// Create a new response definition. To be used e.g. with the ServerEndpoint's Respond builder function.
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

// Set the status code to respond with
func (r *response) StatusCode(statusCode int) *response {
	r.statusCode = statusCode
	return r
}

// Set the body bytes to respond with
func (r *response) Body(body []byte) *response {
	r.body = body
	return r
}

// Set the body (as string) to respond with
func (r *response) BodyString(body string) *response {
	r.body = []byte(body)
	return r
}

// Set a delay, after receiving a request, before sending the response
func (r *response) Delay(delay time.Duration) *response {
	r.delay = delay
	return r
}
