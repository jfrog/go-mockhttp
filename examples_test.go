package mockhttp_test

import (
	"fmt"
	"github.com/jfrog/go-mockhttp"
	"net/http"
)

// Start a mock http server as a test double for other external servers.
// The code under test keeps sending real requests, but instead of sending to the real target, it should send to the mock server.
// The mock server receives requests and responds as configured.
func Example_server() {
	// Start a mock http server with a single endpoint
	server := mockhttp.StartServer(mockhttp.WithEndpoints(
		mockhttp.NewServerEndpoint().
			When(mockhttp.Request().GET("/foo")). // When receiving a GET request to path "/foo"
			Respond(mockhttp.Response())))        // Respond with a default response (status OK)
	// Make sure to close it
	defer server.Close()

	// Issue a request to the server (this is usually done by the code under test...)
	_, _ = http.Get(server.BaseUrl() + "/foo")
	// Check the response and the error

	// Verify that the server got the request as expected
	if err := server.Verify(mockhttp.Request().GET("/foo"), mockhttp.Times(1)); err != nil {
		// Fail the test: t.Errorf("failed expectations: %v", err)
	}
}

// A common use case for the mock http client is to mock transport errors.
func Example_client() {
	// Create a new mock http client with a single endpoint
	// The endpoint matches any request ("When" is not specified), and always returns an error
	client := mockhttp.NewClient(mockhttp.NewClientEndpoint().ReturnError(fmt.Errorf("dummy error")))

	// Provide the client to what ever you test, it will get an error when using the client to send a request
	// In this example we simply use the HTTP client directly...
	_, err := client.HttpClient().Get("http://myhost/foo/bar")

	if err == nil || err.Error() != "Get http://myhost/foo/bar: dummy error" {
		// Fail the test: t.Errorf("error is not as expected: %v", err)
	}
}
