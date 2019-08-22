package mockhttp_test

import (
	"context"
	"fmt"
	"github.com/jfrog/go-mockhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

type subtest struct {
	name     string
	testFunc func(t *testing.T)
}

func TestServer(t *testing.T) {
	tests := []subtest{
		{name: "Simple request and verification", testFunc: subtestSimpleRequest},
		{name: "Run with single server", testFunc: subtestWithSingleServer},
		{name: "Run with multiple servers", testFunc: subtestWithMultipleServers},
	}
	for _, test := range tests {
		t.Run(test.name, test.testFunc)
	}
}

func subtestSimpleRequest(t *testing.T) {
	server1 := mockhttp.StartServer(mockhttp.WithEndpoints(
		mockhttp.NewServerEndpoint().
			When(mockhttp.Request().Method("GET").Path("/svc/foo/bar")).
			Respond(mockhttp.Response().BodyString("hello"))))
	defer server1.Close()
	//First request - expect 200
	assertGetReturns(t, server1.BuildUrl("/svc/foo/bar"), 200, "hello")
	assert.Equal(t, 1, len(server1.AcceptedRequests()), "Unexpected number of accepted requests")
	assert.Equal(t, 0, len(server1.UnmatchedRequests()), "Unexpected number of unmatched requests")
	request := server1.AcceptedRequests()[0]
	assert.Equalf(t, "GET", request.Method, "Recorded request method not as expected. Request:\n%s", request)
	assert.Equalf(t, "", request.BodyAsString(), "Recorded request body not as expected. Request:\n%s", request)
	//Second request - expect 404
	assertGetReturns(t, server1.BuildUrl("/svc/bar"), 404, anyResponseBody)
	assert.Equal(t, 1, len(server1.AcceptedRequests()), "Accepted requests list was not expected to change")
	assert.Equal(t, 1, len(server1.UnmatchedRequests()), "Unexpected number of unmatched requests")
}

func subtestWithSingleServer(t *testing.T) {
	mockhttp.WithServers(mockhttp.ServerSpecs{
		"server1": []mockhttp.ServerEndpoint{
			mockhttp.NewServerEndpoint().
				When(mockhttp.Request().Method("POST").Path("/svc/bar")).
				Respond(mockhttp.Response().BodyString("goodbye")),
		},
	}, func(servers mockhttp.Servers) {
		server1 := servers["server1"]
		res, err := http.Post(server1.BuildUrl("/svc/bar"), "text/plain", strings.NewReader("yoyo"))
		assert.NoError(t, err, "unexpected error")
		assert.Equal(t, 200, res.StatusCode, "unexpected response status")
		assert.Equal(t, "goodbye", string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
		assert.Equal(t, "yoyo", server1.AcceptedRequests()[0].BodyAsString(), "unexpected request body")
		assert.Equal(t, "text/plain", server1.AcceptedRequests()[0].Header.Get("Content-Type"), "unexpected request content-type header")
	})
}

func subtestWithMultipleServers(t *testing.T) {
	mockhttp.WithServers(mockhttp.ServerSpecs{
		"server1": []mockhttp.ServerEndpoint{
			mockhttp.NewServerEndpoint().
				When(mockhttp.Request().Method("GET").Path("/svc1/foo")).
				Respond(mockhttp.Response().BodyString("foo from svc1")),
			mockhttp.NewServerEndpoint().
				When(mockhttp.Request().Method("GET").Path("/svc1/bar")).
				Respond(mockhttp.Response().BodyString("bar from svc1")),
		},
		"server2": []mockhttp.ServerEndpoint{
			mockhttp.NewServerEndpoint().
				When(mockhttp.Request().Method("GET").Path("/svc2/foo")).
				Respond(mockhttp.Response().BodyString("foo from svc2")),
		},
		"server3": []mockhttp.ServerEndpoint{
			mockhttp.NewServerEndpoint().
				When(mockhttp.Request().Method("GET").Path("/svc3/bar")).
				Respond(mockhttp.Response().BodyString("bar from svc3")),
		},
	}, func(servers mockhttp.Servers) {
		server1 := servers["server1"]
		server2 := servers["server2"]
		server3 := servers["server3"]
		assertGetReturns(t, server1.BuildUrl("/svc1/foo"), 200, "foo from svc1")
		assertGetReturns(t, server1.BuildUrl("/svc1/bar"), 200, "bar from svc1")
		assertGetReturns(t, server1.BuildUrl("/svc1/foo"), 200, "foo from svc1")
		assertGetReturns(t, server2.BuildUrl("/svc2/foo"), 200, "foo from svc2")
		assertGetReturns(t, server3.BuildUrl("/svc3/bar"), 200, "bar from svc3")
		assertGetReturns(t, server2.BuildUrl("/svc2/foo"), 200, "foo from svc2")
		assert.Equal(t, 3, len(server1.AcceptedRequests()), "unexpected number of accepted requests for server1")
		assert.Equal(t, 2, len(server2.AcceptedRequests()), "unexpected number of accepted requests for server2")
		assert.Equal(t, 1, len(server3.AcceptedRequests()), "unexpected number of accepted requests for server3")

		assertGetReturns(t, server2.BuildUrl("/svc1/foo"), 404, anyResponseBody)
		assert.Equal(t, 0, len(server1.UnmatchedRequests()), "unexpected number of unmatched requests for server1")
		assert.Equal(t, 1, len(server2.UnmatchedRequests()), "unexpected number of unmatched requests for server2")
		assert.Equal(t, 0, len(server3.UnmatchedRequests()), "unexpected number of unmatched requests for server3")
	})
}

func TestServer_MultipleTestsWithSingleServer(tt *testing.T) {
	server := mockhttp.StartServer()
	defer server.Close()
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "when no endpoint defined return 404 page not found",
			testFunc: func(t *testing.T) {
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 404, "404 page not found")
				assert.Equal(t, 1, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
				assert.Equal(t, "/foo/bar", server.UnmatchedRequests()[0].Path, "unexpected unmatched request path")
			},
		},
		{
			name: "default endpoint accepts all and returns 200 with empty body",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint())
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, "")
				assertGetReturns(t, server.BuildUrl("/bar"), 200, "")
				assertGetReturns(t, server.BuildUrl("/"), 200, "")
				assertGetReturns(t, server.BuildUrl(""), 200, "")
				resp, err := http.Post(server.BuildUrl("/foo"), "foo/bar", strings.NewReader(""))
				assert.NoError(t, err, "unexpected error")
				assert.Equal(t, 200, resp.StatusCode, "unexpected response status code")
				assert.Equal(t, "", string(mockhttp.MustReadAll(t, resp.Body)), "unexpected response body")
				assert.Equal(t, 5, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 0, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
			},
		},
		{
			name: "clear history",
			testFunc: func(t *testing.T) {
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 404, anyResponseBody)
				server.AddEndpoint(mockhttp.NewServerEndpoint())
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, anyResponseBody)
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, anyResponseBody)
				assert.Equal(t, 2, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 1, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
				server.ClearHistory()
				assert.Equal(t, 0, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 0, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, anyResponseBody)
				assert.Equal(t, 1, len(server.AcceptedRequests()), "unexpected number of accepted requests")
			},
		},
		{
			name: "clear server",
			testFunc: func(t *testing.T) {
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 404, anyResponseBody)
				server.AddEndpoint(mockhttp.NewServerEndpoint())
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, anyResponseBody)
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 200, anyResponseBody)
				assert.Equal(t, 2, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 1, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
				server.Clear()
				assert.Equal(t, 0, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 0, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
				assertGetReturns(t, server.BuildUrl("/foo/bar"), 404, anyResponseBody)
				assert.Equal(t, 1, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
			},
		},
		{
			name: "match endpoint by method",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint().When(mockhttp.Request().Method("PATCH")))
				assertGetReturns(t, server.BuildUrl("/foo"), 404, anyResponseBody)
				req, err := http.NewRequest("PATCH", server.BuildUrl("/foo"), nil)
				assert.NoError(t, err, "unexpected error")
				resp, err := http.DefaultClient.Do(req)
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name: "match endpoint by path",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint().When(mockhttp.Request().Path("/foo")))
				assertGetReturns(t, server.BuildUrl("/bar"), 404, anyResponseBody)
				assertGetReturns(t, server.BuildUrl("/foo"), 200, anyResponseBody)
				req, err := http.NewRequest("PATCH", server.BuildUrl("/foo"), nil)
				assert.NoError(t, err, "unexpected error")
				resp, err := http.DefaultClient.Do(req)
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name: "return custom response",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint().Respond(mockhttp.Response().BodyString("hello").StatusCode(202)))
				assertGetReturns(t, server.BuildUrl("/foo"), 202, "hello")
			},
		},
		{
			name: "return custom response with delay",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint().
					Respond(mockhttp.Response().BodyString("hello").StatusCode(202).Delay(200 * time.Millisecond)))
				start := time.Now()
				assertGetReturns(t, server.BuildUrl("/foo"), 202, "hello")
				elapsed := time.Since(start)
				assert.True(t, elapsed >= 200*time.Millisecond, "delay seems to not have been effective")
			},
		},
		{
			name: "return using custom handler",
			testFunc: func(t *testing.T) {
				server.AddEndpoint(mockhttp.NewServerEndpoint().HandleWith(func(response http.ResponseWriter, request *http.Request) {
					response.Header().Set("X-My-Custom-Header", "the-value")
					response.WriteHeader(418)
					_, _ = response.Write([]byte("I'm a teapot!"))
				}))
				resp, err := http.Get(server.BuildUrl("/coffee"))
				assert.NoError(t, err, "unexpected error")
				assert.Equal(t, 418, resp.StatusCode, "unexpected response status code")
				assert.Equal(t, "I'm a teapot!", string(mockhttp.MustReadAll(t, resp.Body)), "unexpected response body")
				assert.Equal(t, "the-value", resp.Header.Get("X-My-Custom-Header"), "unexpected response header value")
			},
		},
		{
			name: "use custom endpoint",
			testFunc: func(t *testing.T) {
				matches := true
				statusCode := 400
				body := "goodbye"
				endpoint := customEndpoint{
					matchesFunc: func(req *http.Request) bool {
						return matches
					},
					serveHttpFunc: func(resp http.ResponseWriter, req *http.Request) {
						resp.WriteHeader(statusCode)
						_, _ = resp.Write([]byte(body))
					},
				}
				server.AddEndpoint(endpoint)
				assertGetReturns(t, server.BuildUrl("/foo"), 400, "goodbye")
				assertGetReturns(t, server.BuildUrl("/foo"), 400, "goodbye")
				matches = false
				assertGetReturns(t, server.BuildUrl("/foo"), 404, anyResponseBody)
				matches = true
				statusCode = 418
				body = "Would you like a cup of tea?"
				assertGetReturns(t, server.BuildUrl("/coffee"), 418, body)
				assert.Equal(t, 3, len(server.AcceptedRequests()), "unexpected number of accepted requests")
				assert.Equal(t, 1, len(server.UnmatchedRequests()), "unexpected number of unmatched requests")
			},
		},
	}
	for _, testCase := range tests {
		tt.Run(testCase.name, testCase.testFunc)
		server.Clear()
	}
}

func TestServer_Verify(t *testing.T) {
	server := mockhttp.StartServer(mockhttp.WithEndpoints(
		mockhttp.NewServerEndpoint().
			When(mockhttp.Request().GET("/foo")).
			Respond(mockhttp.Response())))
	defer server.Close()
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/foo"), mockhttp.Times(0)))
	assertErrorMatches(t, server.Verify(mockhttp.Request().GET("/foo")), regexp.MustCompile("request was called unexpected number of times. expected: 1, actual: 0.*"))
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.Never()))
	assertErrorMatches(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.Once()), regexp.MustCompile("request was called unexpected number of times. expected: 1, actual: 0.*"))

	res, err := http.Get(server.BaseUrl() + "/foo")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	assert.NoError(t, server.Verify(mockhttp.Request().GET("/foo"), mockhttp.Times(1)))
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/foo")))
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.Never()))
	assertErrorMatches(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.AtLeast(1)), regexp.MustCompile("request was called unexpected number of times. expected at least: 1, actual: 0.*"))

	res, err = http.Get(server.BaseUrl() + "/foo")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.Get(server.BaseUrl() + "/bar")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	assert.NoError(t, server.Verify(mockhttp.Request().GET("/foo"), mockhttp.Times(2)))
	assertErrorMatches(t, server.Verify(mockhttp.Request().GET("/foo")), regexp.MustCompile("request was called unexpected number of times. expected: 1, actual: 2.*"))
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.AtLeast(0)))
	assert.NoError(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.AtMost(10)))
	assertErrorMatches(t, server.Verify(mockhttp.Request().GET("/bar"), mockhttp.Times(2)), regexp.MustCompile("request was called unexpected number of times. expected: 2, actual: 1.*"))
}

func TestServer_WaitFor(t *testing.T) {
	server := mockhttp.StartServer(mockhttp.WithEndpoints(
		mockhttp.NewServerEndpoint().
			When(mockhttp.Request().GET("/foo")).
			Respond(mockhttp.Response())))
	defer server.Close()

	// Wait for ends after the expected request is received
	go func() {
		time.Sleep(50 * time.Millisecond)
		res, err := http.Get(server.BaseUrl() + "/foo")
		require.NoError(t, err)
		_ = res.Body.Close()
	}()
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if assert.NoError(t, server.WaitFor(ctx, mockhttp.Request().GET("/foo"))) {
		assertDurationBetween(t, time.Since(start), 50*time.Millisecond, 250*time.Millisecond, "Wait for did not took as expected")
	}

	// Wait for ends with error if timed out waiting for a request
	start = time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if assert.Error(t, server.WaitFor(ctx, mockhttp.Request().GET("/foo"))) {
		assertDurationBetween(t, time.Since(start), 250*time.Millisecond, 500*time.Millisecond, "Wait for did not took as expected")
	}

	// Wait for also checks unmatched requests
	go func() {
		time.Sleep(50 * time.Millisecond)
		res, err := http.Get(server.BaseUrl() + "/bar")
		require.NoError(t, err)
		_ = res.Body.Close()
	}()
	start = time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if assert.NoError(t, server.WaitFor(ctx, mockhttp.Request().GET("/bar"))) {
		assertDurationBetween(t, time.Since(start), 50*time.Millisecond, 250*time.Millisecond, "Wait for did not took as expected")
	}

	// Wait for matches the request correctly (not any request)
	go func() {
		time.Sleep(50 * time.Millisecond)
		res, err := http.Get(server.BaseUrl() + "/foo")
		require.NoError(t, err)
		_ = res.Body.Close()
	}()
	start = time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if assert.Error(t, server.WaitFor(ctx, mockhttp.Request().GET("/bar"))) {
		assertDurationBetween(t, time.Since(start), 250*time.Millisecond, 500*time.Millisecond, "Wait for did not took as expected")
	}
}

func assertDurationBetween(t *testing.T, duration time.Duration, min time.Duration, max time.Duration, msg string, args ...interface{}) bool {
	if duration < min || duration > max {
		return assert.Failf(t, fmt.Sprintf("Duration not in range: \n"+
			"expected: [%v .. %v]\n"+
			"actual  : %v", min, max, duration), msg, args...)
	}
	return true
}

func assertErrorMatches(t *testing.T, err error, msgRegex *regexp.Regexp) bool {
	if assert.Error(t, err) {
		return assert.Regexp(t, msgRegex, err.Error())
	}
	return false
}

type customEndpoint struct {
	matchesFunc   func(req *http.Request) bool
	serveHttpFunc func(resp http.ResponseWriter, req *http.Request)
}

func (e customEndpoint) Matches(req *http.Request) bool {
	return e.matchesFunc(req)
}

func (e customEndpoint) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	e.serveHttpFunc(resp, req)
}

const (
	anyResponseStatus int    = 0
	anyResponseBody   string = "@@@@@_ANY_RESPONSE_BODY_@@@@@"
)

func assertGetReturns(t *testing.T, url string, expectedStatus int, expectedBody string) {
	res, err := http.Get(url)
	assert.NoError(t, err, "unexpected error")
	if expectedStatus > anyResponseStatus {
		assert.Equal(t, expectedStatus, res.StatusCode, "Response status not as expected")
	}
	if expectedBody != anyResponseBody {
		assert.Equal(t, expectedBody, string(mockhttp.MustReadAll(t, res.Body)), "unexpected response body")
	}
}
