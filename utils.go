package mockhttp

import (
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"testing"
)

type Servers map[string]*Server
type ServerSpecs map[string][]ServerEndpoint
type TestWithMockServers func(servers Servers)

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

func MustReadAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	data, err := ioutil.ReadAll(r)
	require.NoError(t, err, "could not read all bytes")
	return data
}
