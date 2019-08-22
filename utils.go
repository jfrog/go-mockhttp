package mockhttp

import (
	"io"
	"io/ioutil"
	"testing"
)

// Map of mock http servers
type Servers map[string]*Server

// Map of server specifications, used for specifying multiple servers to start
type ServerSpecs map[string][]ServerEndpoint

// Test function which receives a collection of running mock http servers
type TestWithMockServers func(servers Servers)

// Helper function to run a test with multiple mock http servers.
//
// Providing mock http server specifications, this function will make sure to start all servers, run the test, and close
// all servers after the test. This way the test function can focus on the test itself instead of managing the mock http
// servers. The test function receives a map of servers (already configured and running), matching the map of server
// specifications.
func WithServers(serverSpecs ServerSpecs, test TestWithMockServers) {
	servers := make(Servers)
	defer func() {
		for _, v := range servers {
			v.Close()
		}
	}()
	for k, v := range serverSpecs {
		servers[k] = StartServer(WithName(k), WithEndpoints(v...))
	}
	test(servers)
}

// Helper utility function to read all bytes of a given reader. Will fail the test in case of an error while reading.
func MustReadAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("could not read all bytes: %v", err)
	}
	return data
}
