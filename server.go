package mockhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

// Functional option for configuring a mock http server
type ServerOpt func(*Server)

// Set the name of the mock http server.
//
// Used mainly for logging, has no real functional purpose. Set to "anonymous" if not explicitly set.
func WithName(name string) ServerOpt {
	return func(s *Server) {
		s.name = name
	}
}

// Set TLS configuration, to start the mock http server with TLS enabled.
//
// A server is started without TLS by default if not explicitly set.
func WithTls(config *tls.Config) ServerOpt {
	return func(s *Server) {
		s.tlsConfig = config
	}
}

// Set the endpoints the server shall handle
func WithEndpoints(endpoints ...ServerEndpoint) ServerOpt {
	return func(s *Server) {
		s.endpoints = endpoints
	}
}

func defaultServer() *Server {
	return &Server{
		name:            "anonymous",
		requestRecorder: newRequestRecorder(),
	}
}

// Start a new mock http server
//
// The server is configured using the provided functional options.
// The defaults are:
//   - Name: "anonymous"
//   - TLS disabled
//   - No handled endpoints - all requests return 404
//
// Make sure to close the server when done. A common practice is to use:
//   server := StartServer() // Configure as needed
//   defer server.Close()
func StartServer(opts ...ServerOpt) *Server {
	mockSvr := defaultServer()
	for _, opt := range opts {
		opt(mockSvr)
	}

	mockSvr.server = httptest.NewUnstartedServer(&httpHandler{mockSvr: mockSvr})
	if mockSvr.tlsConfig != nil {
		mockSvr.server.TLS = mockSvr.tlsConfig
		mockSvr.server.StartTLS()
	} else {
		mockSvr.server.Start()
	}

	if tcpAddr, ok := mockSvr.server.Listener.Addr().(*net.TCPAddr); ok {
		mockSvr.Port = tcpAddr.Port
	} else {
		panic(fmt.Errorf("unexpected state - listener address is not a TCP address: %v", mockSvr.server.Listener.Addr()))
	}
	fmt.Printf("Mock server started: %s\n", mockSvr)
	return mockSvr
}

// Mock http server
type Server struct {
	Port            int
	name            string
	server          *httptest.Server
	endpoints       []ServerEndpoint
	requestRecorder *requestRecorder
	tlsConfig       *tls.Config
}

// Close (shutdown) the server
func (mockSvr *Server) Close() {
	fmt.Printf("Closing mock server '%s'.\n", mockSvr.name)
	mockSvr.server.Close()
}

// The base URL of this server
//
// The URL is constructed based on whether TLS is enabled ("http" or "https") and on the port the server started with.
// An example base URL would be:
//   http://localhost:54756
func (mockSvr *Server) BaseUrl() string {
	scheme := "http"
	if mockSvr.tlsConfig != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://localhost:%d", scheme, mockSvr.Port)
}

// Build a URL based on the server's base URL and the given path
//
// For example:
//   url := server.BuildUrl("/path/to/something")
//   // url == "http://localhost:54756/path/to/something"
func (mockSvr *Server) BuildUrl(path string) string {
	return fmt.Sprintf("%s%s", mockSvr.BaseUrl(), path)
}

// Add an endpoint to this server
func (mockSvr *Server) AddEndpoint(endpoint ServerEndpoint) {
	mockSvr.endpoints = append(mockSvr.endpoints, endpoint)
}

// Clear request history and remove all endpoints defined for this server
//
// Helpful when reusing the same mock http server for multiple tests, to make sure a clean start
func (mockSvr *Server) Clear() {
	mockSvr.endpoints = []ServerEndpoint{}
	mockSvr.ClearHistory()
}

// Get all requests which got to this server and were handled by one of the defined endpoints
func (mockSvr *Server) AcceptedRequests() []recordedRequest {
	return mockSvr.requestRecorder.AcceptedRequests()
}

// Get all requests which got to this server but did not match any of the defined endpoints
func (mockSvr *Server) UnmatchedRequests() []recordedRequest {
	return mockSvr.requestRecorder.UnmatchedRequests()
}

// Clean all the request history recorded by this server
func (mockSvr *Server) ClearHistory() {
	mockSvr.requestRecorder.ClearHistory()
}

// Verify requests received by this server. Requires a request matcher to specify which requests to check and optionally
// verify options (e.g. how many times, etc.). If it does not match the expectation, an error is returned, otherwise
// returns nil.
//
// For example:
//   // verify that the server got a GET request with path "/foo" exactly twice
//   err := server.Verify(Request().GET("/foo"), Times(2)
//   if err != nil {
//   	t.Errorf("%v", err)
//   }
func (mockSvr *Server) Verify(matcher *requestMatcher, opts ...verifyOpt) error {
	return newVerifier(matcher, opts...).verifyRequests(mockSvr.requestRecorder)
}

// Wait for a request (matching the given matcher) to be received by the server, no matter if an matching endpoint is
// defined. The provided context can be used e.g. for setting a timeout. Returns an error e.g. when waiting has timed out.
//
// For example:
//   // Wait for any request with timeout of 5 seconds
//   ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
//   defer cancel()
//   err := server.WaitFor(ctx, Request())
//
func (mockSvr *Server) WaitFor(ctx context.Context, matcher *requestMatcher) error {
	verifier := newVerifier(matcher)
	countRequests := func() int {
		accepted := verifier.countRequests(mockSvr.requestRecorder.AcceptedRequests())
		unmatched := verifier.countRequests(mockSvr.requestRecorder.UnmatchedRequests())
		return accepted + unmatched
	}
	initialCount := countRequests()
	for {
		if countRequests() > initialCount {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (mockSvr *Server) String() string {
	return fmt.Sprintf("'%s' - base URL: %s", mockSvr.name, mockSvr.BaseUrl())
}

type httpHandler struct {
	mockSvr *Server
}

func (h *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	for _, endpoint := range h.mockSvr.endpoints {
		if endpoint.Matches(request) {
			h.mockSvr.requestRecorder.recordAcceptedRequest(request)
			endpoint.ServeHTTP(response, request)
			return
		}
	}
	h.mockSvr.requestRecorder.recordUnmatchedRequest(request)
	response.Header().Set("Content-Type", "text/plain")
	response.WriteHeader(404)
	response.Write([]byte("404 page not found"))
}
