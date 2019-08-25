package mockhttp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func TestRecordedRequest(t *testing.T) {
	tests := []struct {
		method   string
		url      string
		header   http.Header
		body     string
		expected string
	}{
		{
			method: "GET",
			url:    "http://host/foo/bar?q1=2&q2=3",
			header: http.Header{"X-Foo": {"foo"}, "X-Bar": {"bar", "baz"}},
			expected: `GET /foo/bar?q1=2&q2=3
X-Bar: bar
X-Bar: baz
X-Foo: foo

`,
		},
		{
			method: "POST",
			url:    "http://host:8081/bar/foo",
			header: make(http.Header),
			body:   "hello",
			expected: `POST /bar/foo
hello
`,
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s %s - header:%s;body:'%s'", tt.method, tt.url, tt.header, tt.body)
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header = tt.header

			recReq := newRecordedRequest(req)

			assert.Equal(t, req.Method, recReq.Method, "unexpected method")
			assert.Equal(t, req.URL.Path, recReq.Path, "unexpected url path")
			assert.Equal(t, string(MustReadAll(t, req.Body)), string(recReq.Body), "unexpected body")
			assert.Equal(t, tt.body, recReq.BodyAsString(), "unexpected body as string")
			assert.Equal(t, req.Header, recReq.Header, "unexpected headers")
			assert.Equal(t, req.URL.Query(), recReq.Query, "unexpected query params")
			assert.Equal(t, tt.expected, strings.Replace(recReq.String(), "\r", "", -1), "unexpected recorded request as string")

			newReq := recReq.toHttpRequest()
			assert.Equal(t, tt.method, newReq.Method, "unexpected method in new http request")
			assert.Equal(t, req.URL.Path, newReq.URL.Path, "unexpected url path in new http request")
			assert.Equal(t, req.URL.Query(), newReq.URL.Query(), "unexpected query params in new http request")
			assert.Equal(t, tt.header, newReq.Header, "unexpected header in new http request")
			assert.Equal(t, tt.body, string(MustReadAll(t, newReq.Body)), "unexpected body in new http request")
		})
	}
}

func TestRequestRecorder(t *testing.T) {
	recorder := newRequestRecorder()
	assert.Equal(t, 0, len(recorder.AcceptedRequests()), "unexpected number of accepted requests")
	assert.Equal(t, 0, len(recorder.UnmatchedRequests()), "unexpected number of unmatched requests")
	assert.Equal(t, "Accepted:\nUnmatched:\n", recorder.String(), "unexpected recorder as string")

	req, err := http.NewRequest("GET", "http://host/foo/bar", nil)
	require.NoError(t, err)
	recorder.recordAcceptedRequest(req)
	assert.Equal(t, 1, len(recorder.AcceptedRequests()), "unexpected number of accepted requests")
	assert.Equal(t, 0, len(recorder.UnmatchedRequests()), "unexpected number of unmatched requests")

	req, err = http.NewRequest("DELETE", "http://host/bar", nil)
	require.NoError(t, err)
	recorder.recordUnmatchedRequest(req)
	assert.Equal(t, 1, len(recorder.AcceptedRequests()), "unexpected number of accepted requests")
	assert.Equal(t, 1, len(recorder.UnmatchedRequests()), "unexpected number of unmatched requests")

	req, err = http.NewRequest("POST", "http://host/foo/baz", strings.NewReader("goodbye"))
	require.NoError(t, err)
	recorder.recordAcceptedRequest(req)
	require.Equal(t, 2, len(recorder.AcceptedRequests()), "unexpected number of accepted requests")
	require.Equal(t, 1, len(recorder.UnmatchedRequests()), "unexpected number of unmatched requests")
	assert.Equal(t, "GET", recorder.AcceptedRequests()[0].Method, "unexpected method for 1st accepted request")
	assert.Equal(t, "/foo/bar", recorder.AcceptedRequests()[0].Path, "unexpected path for 1st accepted request")
	assert.Equal(t, "POST", recorder.AcceptedRequests()[1].Method, "unexpected method for 2nd accepted request")
	assert.Equal(t, "/foo/baz", recorder.AcceptedRequests()[1].Path, "unexpected path for 2nd accepted request")
	assert.Equal(t, "DELETE", recorder.UnmatchedRequests()[0].Method, "unexpected method for 1st unmatched request")
	assert.Equal(t, "/bar", recorder.UnmatchedRequests()[0].Path, "unexpected path for 1st unmatched request")
	assert.Equal(t, "Accepted:\n   1: GET /foo/bar\n   2: POST /foo/baz\nUnmatched:\n   1: DELETE /bar\n", recorder.String(), "unexpected recorder as string")

	recorder.ClearHistory()
	assert.Equal(t, 0, len(recorder.AcceptedRequests()), "unexpected number of accepted requests after history cleanup")
	assert.Equal(t, 0, len(recorder.UnmatchedRequests()), "unexpected number of unmatched requests after history cleanup")
}
