package mockhttp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServerEndpoint_responseAsHandler(t *testing.T) {
	const delayUnit = 50 * time.Millisecond
	tests := []struct {
		name     string
		response *response
	}{
		{
			name: "no headers, no delay, no body",
			response: Response().
				StatusCode(http.StatusOK),
		},
		{
			name: "with headers, with delay, with body",
			response: Response().
				StatusCode(http.StatusNotFound).
				Header("X-Foo", "foo").
				Header("X-Bar", "bar1", "bar2").
				BodyString("How long?").
				Delay(2 * delayUnit),
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := responseAsHandler(testCase.response)

			start := time.Now()
			handler(recorder, &http.Request{})
			elapsed := time.Since(start)

			assert.Equal(t, testCase.response.statusCode, recorder.Code)
			assert.Equal(t, testCase.response.header, recorder.Header())
			assert.Equal(t, string(testCase.response.body), string(recorder.Body.Bytes()))
			assert.True(t, elapsed > testCase.response.delay-delayUnit, fmt.Sprintf("response delay was less than expected: %v (expected: %v)", elapsed, testCase.response.delay))
			assert.True(t, elapsed < testCase.response.delay+delayUnit, fmt.Sprintf("response delay was more than expected: %v (expected: %v)", elapsed, testCase.response.delay))
		})
	}
}
