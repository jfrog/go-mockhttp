package mockhttp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestRequestMatcher_Default(t *testing.T) {
	requests := []http.Request{
		{},
		{Method: "GET"},
		{Method: "PUT"},
		{URL: &url.URL{Path: "/foo"}},
		{URL: &url.URL{Path: "/bar/foo"}},
		{Header: http.Header{"X-Foo": []string{"hello"}}},
		{Method: "POST", URL: &url.URL{Path: "/bar/foo"}, Header: http.Header{"X-Foo": []string{"hello"}}},
	}
	for _, req := range requests {
		assert.True(t, Request().matches(&req), "request does not match", req)
	}
}

func TestRequestMatcher_GET(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Method: "GET", URL: &url.URL{Path: "/bar/foo"}},
			want:    true,
		},
		{
			request: http.Request{Method: "GET", URL: &url.URL{Path: "/foo"}},
			want:    false,
		},
		{
			request: http.Request{Method: "PUT", URL: &url.URL{Path: "/bar/foo"}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().GET("/bar/foo").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_POST(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Method: "POST", URL: &url.URL{Path: "/bar/foo"}},
			want:    true,
		},
		{
			request: http.Request{Method: "POST", URL: &url.URL{Path: "/foo"}},
			want:    false,
		},
		{
			request: http.Request{Method: "PUT", URL: &url.URL{Path: "/bar/foo"}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().POST("/bar/foo").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_PUT(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Method: "PUT", URL: &url.URL{Path: "/bar/foo"}},
			want:    true,
		},
		{
			request: http.Request{Method: "PUT", URL: &url.URL{Path: "/foo"}},
			want:    false,
		},
		{
			request: http.Request{Method: "POST", URL: &url.URL{Path: "/bar/foo"}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().PUT("/bar/foo").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_DELETE(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Method: "DELETE", URL: &url.URL{Path: "/bar/foo"}},
			want:    true,
		},
		{
			request: http.Request{Method: "DELETE", URL: &url.URL{Path: "/foo"}},
			want:    false,
		},
		{
			request: http.Request{Method: "PUT", URL: &url.URL{Path: "/bar/foo"}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().DELETE("/bar/foo").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_Method(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Method: "PATCH"},
			want:    true,
		},
		{
			request: http.Request{Method: "DELETE"},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().Method("PATCH").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_Path(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{URL: &url.URL{Path: "/foo/bar"}},
			want:    true,
		},
		{
			request: http.Request{URL: &url.URL{Path: "/foo"}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().Path("/foo/bar").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_PathMatches(t *testing.T) {
	tests := []struct {
		request http.Request
		regex   *regexp.Regexp
		want    bool
	}{
		{
			request: http.Request{URL: &url.URL{Path: "/foo/bar"}},
			regex:   regexp.MustCompile("^/foo/bar$"),
			want:    true,
		},
		{
			request: http.Request{URL: &url.URL{Path: "/foo/bar"}},
			regex:   regexp.MustCompile("/foo/.+"),
			want:    true,
		},
		{
			request: http.Request{URL: &url.URL{Path: "/foo"}},
			regex:   regexp.MustCompile("hello"),
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().PathMatches(testCase.regex).matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_Header(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"hello"}}},
			want:    true,
		},
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"hello world"}}},
			want:    false,
		},
		{
			request: http.Request{},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().Header("X-Foo", "hello").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_HeaderMatches(t *testing.T) {
	tests := []struct {
		request http.Request
		regex   *regexp.Regexp
		want    bool
	}{
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"hello"}}},
			regex:   regexp.MustCompile("hello"),
			want:    true,
		},
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"hello"}}},
			regex:   regexp.MustCompile("h.+o"),
			want:    true,
		},
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"hello world"}}},
			regex:   regexp.MustCompile(".* world"),
			want:    true,
		},
		{
			request: http.Request{Header: http.Header{"X-Bar": []string{"hello"}}},
			regex:   regexp.MustCompile("hello"),
			want:    false,
		},
		{
			request: http.Request{},
			regex:   regexp.MustCompile("hello"),
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().HeaderMatches("X-Foo", testCase.regex).matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}

func TestRequestMatcher_NoHeader(t *testing.T) {
	tests := []struct {
		request http.Request
		want    bool
	}{
		{
			request: http.Request{Header: http.Header{"X-Bar": []string{"hello"}}},
			want:    true,
		},
		{
			request: http.Request{},
			want:    true,
		},
		{
			request: http.Request{Header: http.Header{"X-Foo": []string{"world"}}},
			want:    false,
		},
	}
	for _, testCase := range tests {
		assert.Equalf(t, testCase.want, Request().NoHeader("X-Foo").matches(&testCase.request), "request match not as expected. want match: %b, request: %+v", testCase.want, testCase.request)
	}
}
