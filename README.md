# go-mockhttp

Package for mocking HTTP servers and clients. 

## Main Features
  
It provides the following main features:
1. Mock using real HTTP server with simply defined endpoints and expected behavior
1. Verify expectations
1. Simulate server faults, such as response delays
1. Simulate transport errors, such as connection failures

## Usage

A common use case is to start a mock HTTP server in your test, define endpoints, call your code which is expected to 
send requests to an HTTP server, and verify the expectations.

```go
// Start a mock http server with a single endpoint which matches a request and returns a response
server := mockhttp.StartServer(mockhttp.WithEndpoints(
	mockhttp.NewServerEndpoint().
		When(mockhttp.Request().GET("/foo")). // When received request matches: GET /foo (with any headers)
		Respond(mockhttp.Response())))        // Respond with default response (status 200 OK, empty body, no headers)
// Make sure to close it
defer server.Close()

// Issue a request to the server (this is usually done by the code under test...)
res, err := http.Get(server.BaseUrl() + "/foo")
// check the error

// Verify that the server got the request as expected
if err := server.Verify(mockhttp.Request().GET("/foo"), mockhttp.Times(1)); err != nil {
	t.Errorf("failed expectations: %v", err)
}
```

The above is a very basic and simple use case. For mode details and usage options, see the package's documentation.