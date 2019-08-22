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

func Response() *response {
	return &response{
		statusCode: http.StatusOK,
		body:       []byte{},
		header:     http.Header{},
	}
}

func (r *response) StatusCode(statusCode int) *response {
	r.statusCode = statusCode
	return r
}

func (r *response) Body(body []byte) *response {
	r.body = body
	return r
}

func (r *response) BodyString(body string) *response {
	r.body = []byte(body)
	return r
}

func (r *response) Delay(delay time.Duration) *response {
	r.delay = delay
	return r
}
