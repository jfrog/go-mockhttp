package mockhttp_test

import (
	"fmt"
	"github.com/jfrog/go-mockhttp"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestClient(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name:     "DefaultEndpoint",
			testFunc: subtest_DefaultEndpoint,
		},
		{
			name:     "AnyRequest_CustomResponse",
			testFunc: subtest_AnyRequest_CustomResponse,
		},
		{
			name:     "AnyRequest_CustomHandler",
			testFunc: subtest_AnyRequest_CustomHandler,
		},
		{
			name:     "MatchRequests",
			testFunc: subtest_MatchRequests,
		},
		{
			name:     "EmptyClient",
			testFunc: subtest_EmptyClient,
		},
		{
			name:     "TransportError",
			testFunc: subtest_TransportError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, testCase.testFunc)
	}
}

func subtest_DefaultEndpoint(t *testing.T) {
	client := mockhttp.NewClient(mockhttp.NewClientEndpoint())
	assertClientRecordedRequestCount(t, client, 0, 0)
	res, err := client.HttpClient().Get("http://myhost/foo/bar")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "unexpected response status code")
	assert.Equal(t, []byte{}, mockhttp.MustReadAll(t, res.Body), "unexpected response body")
	assert.NotNil(t, res.Header, "response header should not be nil")
	assertClientRecordedRequestCount(t, client, 1, 0)
}

func subtest_AnyRequest_CustomResponse(t *testing.T) {
	client := mockhttp.NewClient(mockhttp.NewClientEndpoint().
		Respond(mockhttp.Response().StatusCode(http.StatusTeapot).BodyString("I'm a teapot!")))
	res, err := client.HttpClient().Get("http://myhost/coffee")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, res.StatusCode, "unexpected response status code")
	assert.Equal(t, "I'm a teapot!", string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
	assertClientRecordedRequestCount(t, client, 1, 0)
}

func subtest_AnyRequest_CustomHandler(t *testing.T) {
	client := mockhttp.NewClient(mockhttp.NewClientEndpoint().HandleWith(func(request *http.Request) (response *http.Response, e error) {
		return nil, fmt.Errorf("custom error")
	}))
	res, err := client.HttpClient().Get("http://myhost/foo")
	assert.Nil(t, res, "response was not expected")
	assert.EqualError(t, err, "Get http://myhost/foo: custom error", "expected an error with a specific message")
}

func subtest_MatchRequests(t *testing.T) {
	client := mockhttp.NewClient(
		mockhttp.NewClientEndpoint().
			When(mockhttp.Request().Method("GET").Path("/coffee")).
			Respond(mockhttp.Response().StatusCode(http.StatusTeapot).BodyString("I'm a teapot!")),
		mockhttp.NewClientEndpoint().
			When(mockhttp.Request().Method("POST").Path("/foo/bar")).
			Respond(mockhttp.Response().StatusCode(http.StatusCreated).BodyString("done")),
	)
	assertClientRecordedRequestCount(t, client, 0, 0)

	res, err := client.HttpClient().Post("http://myhost/foo/bar", "plain/text", strings.NewReader("foo"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode, "unexpected response status code")
	assert.Equal(t, "done", string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
	assert.Equal(t, "POST", client.AcceptedRequests()[0].Method, "unexpected recorded request method")
	assert.Equal(t, "/foo/bar", client.AcceptedRequests()[0].Path, "unexpected recorded request path")

	res, err = client.HttpClient().Get("http://myhost/coffee")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, res.StatusCode, "unexpected response status code")
	assert.Equal(t, "I'm a teapot!", string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
	assert.Equal(t, "GET", client.AcceptedRequests()[1].Method, "unexpected recorded request method")
	assert.Equal(t, "/coffee", client.AcceptedRequests()[1].Path, "unexpected recorded request path")

	res, err = client.HttpClient().Head("http://myhost/foo/bar")
	assert.NoError(t, err)
	assertNotImplementedResponse(t, res)
	assert.Equal(t, "HEAD", client.UnmatchedRequests()[0].Method, "unexpected recorded request method")
	assert.Equal(t, "/foo/bar", client.UnmatchedRequests()[0].Path, "unexpected recorded request path")

	res, err = client.HttpClient().Post("http://myhost/foo/bar", "plain/text", strings.NewReader("bar"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode, "unexpected response status code")
	assert.Equal(t, "done", string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
	assert.Equal(t, "POST", client.AcceptedRequests()[2].Method, "unexpected recorded request method")
	assert.Equal(t, "/foo/bar", client.AcceptedRequests()[2].Path, "unexpected recorded request path")

	assertClientRecordedRequestCount(t, client, 3, 1)

	client.ClearHistory()
	assertClientRecordedRequestCount(t, client, 0, 0)
}

func subtest_EmptyClient(t *testing.T) {
	client := mockhttp.NewClient()
	res, err := client.HttpClient().Get("http://myhost/foo/bar")
	assert.NoError(t, err)
	assertNotImplementedResponse(t, res)
	assertClientRecordedRequestCount(t, client, 0, 1)
}

func subtest_TransportError(t *testing.T) {
	client := mockhttp.NewClient(mockhttp.NewClientEndpoint().ReturnError(fmt.Errorf("dummy error")))
	res, err := client.HttpClient().Get("http://myhost/foo/bar")
	assert.Nil(t, res, "response was not expected")
	assert.EqualError(t, err, "Get http://myhost/foo/bar: dummy error", "expected an error with a specific message")
}

func assertNotImplementedResponse(t *testing.T, res *http.Response) {
	req := res.Request
	assert.Equal(t, http.StatusNotImplemented, res.StatusCode, "unexpected response status code")
	assert.Equal(t, fmt.Sprintf("Unmatched request: %s %s", req.Method, req.URL), string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
}

func assertClientRecordedRequestCount(t *testing.T, c *mockhttp.Client, expectedAccepted int, expectedUnmatched int) {
	assert.Equal(t, expectedAccepted, len(c.AcceptedRequests()), "Unexpected number of accepted requests")
	assert.Equal(t, expectedUnmatched, len(c.UnmatchedRequests()), "Unexpected number of unmatched requests")
}
