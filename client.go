package mockhttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func NewClient(endpoints ...ClientEndpoint) *Client {
	client := Client{}
	client.endpoints = endpoints
	client.requestRecorder = newRequestRecorder()
	client.httpClient = &http.Client{
		Transport: &roundTripper{client: &client},
	}
	return &client
}

type Client struct {
	httpClient      *http.Client
	endpoints       []ClientEndpoint
	requestRecorder *requestRecorder
}

func (c *Client) HttpClient() *http.Client {
	return c.httpClient
}

func (c *Client) AcceptedRequests() []recordedRequest {
	return c.requestRecorder.AcceptedRequests()
}

func (c *Client) UnmatchedRequests() []recordedRequest {
	return c.requestRecorder.UnmatchedRequests()
}

func (c *Client) ClearHistory() {
	c.requestRecorder.ClearHistory()
}

type roundTripper struct {
	client *Client
}

func (r *roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if r.client.endpoints != nil {
		for _, endpoint := range r.client.endpoints {
			if endpoint.Matches(request) {
				r.client.requestRecorder.recordAcceptedRequest(request)
				return endpoint.RoundTrip(request)
			}
		}
	}
	r.client.requestRecorder.recordUnmatchedRequest(request)
	return unmatchedRequestResponse(request), nil
}

func unmatchedRequestResponse(request *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusNotImplemented,
		Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("Unmatched request: %s %s", request.Method, request.URL))),
		Header:     http.Header{},
		Request:    request,
	}
}
