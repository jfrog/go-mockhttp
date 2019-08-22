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

type ServerOpt func(*Server)

func WithName(name string) ServerOpt {
	return func(s *Server) {
		s.name = name
	}
}

func WithTls(config *tls.Config) ServerOpt {
	return func(s *Server) {
		s.tlsConfig = config
	}
}

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

type Server struct {
	Port            int
	name            string
	server          *httptest.Server
	endpoints       []ServerEndpoint
	requestRecorder *requestRecorder
	tlsConfig       *tls.Config
}

func (mockSvr *Server) Close() {
	fmt.Printf("Closing mock server '%s'.\n", mockSvr.name)
	mockSvr.server.Close()
}

func (mockSvr *Server) BaseUrl() string {
	return fmt.Sprintf("http://localhost:%d", mockSvr.Port)
}

func (mockSvr *Server) BuildUrl(path string) string {
	return fmt.Sprintf("%s%s", mockSvr.BaseUrl(), path)
}

func (mockSvr *Server) AddEndpoint(endpoint ServerEndpoint) {
	mockSvr.endpoints = append(mockSvr.endpoints, endpoint)
}

func (mockSvr *Server) Clear() {
	mockSvr.endpoints = []ServerEndpoint{}
	mockSvr.ClearHistory()
}

func (mockSvr *Server) AcceptedRequests() []recordedRequest {
	return mockSvr.requestRecorder.AcceptedRequests()
}

func (mockSvr *Server) UnmatchedRequests() []recordedRequest {
	return mockSvr.requestRecorder.UnmatchedRequests()
}

func (mockSvr *Server) ClearHistory() {
	mockSvr.requestRecorder.ClearHistory()
}

func (mockSvr *Server) Verify(matcher *requestMatcher, opts ...verifyOpt) error {
	return newVerifier(matcher, opts...).verifyRequests(mockSvr.requestRecorder)
}

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
