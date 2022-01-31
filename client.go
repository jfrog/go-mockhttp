package mockhttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// NewClient creates a new mock http client with a list of client endpoints it should handle
//
// A common use case for this mock http client is to mock transport errors. For example:
//   // create a new mock http client which returns a given error
//   client := NewClient(NewClientEndpoint().ReturnError(fmt.Errorf("dummy error")))
//
//   //provide the client to what ever you test, it will get an error when using the client to send a request
//   result, err := mylib.CallSomethingUsingClient(client.HttpClient())
//
//	 // assert the above call behaves as expected, e.g. returns an error
func NewClient(endpoints ...ClientEndpoint) *Client {
	client := Client{}
	client.endpoints = endpoints
	client.requestRecorder = newRequestRecorder()
	client.httpClient = &http.Client{
		Transport: &roundTripper{client: &client},
	}
	return &client
}

// Client is a mock http client
type Client struct {
	httpClient      *http.Client
	endpoints       []ClientEndpoint
	requestRecorder *requestRecorder
}

// HttpClient returns the actual http client, to be used by tests
func (c *Client) HttpClient() *http.Client {
	return c.httpClient
}

// AcceptedRequests returns all requests which this client received and were handled by one of the defined endpoints
func (c *Client) AcceptedRequests() []recordedRequest {
	return c.requestRecorder.AcceptedRequests()
}

// UnmatchedRequests returns all requests which this client received but did not match any of the defined endpoints
func (c *Client) UnmatchedRequests() []recordedRequest {
	return c.requestRecorder.UnmatchedRequests()
}

// ClearHistory cleans all the request history recorded by this client
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
