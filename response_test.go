package mockhttp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestResponse_DefaultValues(t *testing.T) {
	res := Response()
	assert.Equal(t, http.StatusOK, res.statusCode)
	assert.Equal(t, http.Header{}, res.header)
	assert.Equal(t, []byte{}, res.body)
	assert.Equal(t, time.Duration(0), res.delay)
}

func TestResponse_CustomValues(t *testing.T) {
	res := Response().
		StatusCode(http.StatusTeapot).
		Header("X-Foo", "will be overridden").
		Header("X-Foo", "foo").
		Header("X-Bar", "bar1", "bar2", "bar3").
		Body([]byte("I'm a teapot")).
		Delay(time.Second)
	assert.Equal(t, http.StatusTeapot, res.statusCode)
	assert.Equal(t, http.Header{"X-Foo": []string{"foo"}, "X-Bar": []string{"bar1", "bar2", "bar3"}}, res.header)
	assert.Equal(t, []byte("I'm a teapot"), res.body)
	assert.Equal(t, time.Second, res.delay)
}

func TestResponse_BodyString(t *testing.T) {
	res := Response().BodyString("Hello World!")
	assert.Equal(t, []byte("Hello World!"), res.body)
}
