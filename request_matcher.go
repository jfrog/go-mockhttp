package mockhttp

import (
	"fmt"
	"net/http"
	"regexp"
)

type requestMatcherFunc func(request *http.Request) bool

type requestMatcher struct {
	requestMatchers []requestMatcherFunc
	description     string
}

// Create a new request matcher. By default it matches everything. Use configuration methods to narrow down what matches
// and what not. All limitations are handled with AND.
//
// For example:
//   Request().
//   	Method("DELETE").
//   	Path("/foo").
//   	Header("Content-Type", "application/json")
func Request() *requestMatcher {
	return &requestMatcher{}
}

func (m *requestMatcher) matches(request *http.Request) bool {
	if m.requestMatchers != nil {
		for _, matches := range m.requestMatchers {
			if !matches(request) {
				return false
			}
		}
	}
	return true
}

// Match the given http method (exact match)
func (m *requestMatcher) Method(method string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Method(%s)", method), func(request *http.Request) bool {
		return request.Method == method
	})
	return m
}

// Match the given path (exact match)
func (m *requestMatcher) Path(path string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Path(%s)", path), func(request *http.Request) bool {
		return request.URL.Path == path
	})
	return m
}

// Use the given regular expression to match the request path
func (m *requestMatcher) PathMatches(path *regexp.Regexp) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("PathMatches(%s)", path), func(request *http.Request) bool {
		return path.MatchString(request.URL.Path)
	})
	return m
}

// Match a GET request with the given path.
//   GET("/foo")
// Which is a shortcut for:
//   Method("GET").Path("/foo")
func (m *requestMatcher) GET(path string) *requestMatcher {
	return m.Method("GET").Path(path)
}

// Match a POST request with the given path.
//   POST("/foo")
// Which is a shortcut for:
//   Method("POST").Path("/foo")
func (m *requestMatcher) POST(path string) *requestMatcher {
	return m.Method("POST").Path(path)
}

// Match a PUT request with the given path.
//   PUT("/foo")
// Which is a shortcut for:
//   Method("PUT").Path("/foo")
func (m *requestMatcher) PUT(path string) *requestMatcher {
	return m.Method("PUT").Path(path)
}

// Match a DELETE request with the given path.
//   DELETE("/foo")
// Which is a shortcut for:
//   Method("DELETE").Path("/foo")
func (m *requestMatcher) DELETE(path string) *requestMatcher {
	return m.Method("DELETE").Path(path)
}

// Match a request with the given header key-value pair (exact match)
func (m *requestMatcher) Header(key string, value string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Header(%s: %s)", key, value), func(request *http.Request) bool {
		return request.Header.Get(key) == value
	})
	return m
}

// Match a request with the given header key-value pair, the value is evaluated using the given regular expression
func (m *requestMatcher) HeaderMatches(key string, value *regexp.Regexp) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("HeaderMatches(%s: %s)", key, value), func(request *http.Request) bool {
		return value.MatchString(request.Header.Get(key))
	})
	return m
}

// Match a request only if it does not have a header with the given key
func (m *requestMatcher) NoHeader(key string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("NoHeader(%s)", key), func(request *http.Request) bool {
		return request.Header.Get(key) == ""
	})
	return m
}

// Match a request with the given query key-value pair (exact match)
func (m *requestMatcher) Query(key string, value string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Query(%s: %s)", key, value), func(request *http.Request) bool {
		return request.URL != nil && request.URL.Query().Get(key) == value
	})
	return m
}

// Match a request with the given query key-value pair, the value is evaluated using the given regular expression
func (m *requestMatcher) QueryMatches(key string, value *regexp.Regexp) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("QueryMatches(%s: %s)", key, value), func(request *http.Request) bool {
		return request.URL != nil && value.MatchString(request.URL.Query().Get(key))
	})
	return m
}

// Match a request only if it does not have a query with the given key
func (m *requestMatcher) NoQuery(key string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("NoQuery(%s)", key), func(request *http.Request) bool {
		return request.URL == nil || request.URL.Query().Get(key) == ""
	})
	return m
}

func (m *requestMatcher) appendMatcher(desc string, matcher requestMatcherFunc) {
	if len(m.description) > 0 {
		m.description = m.description + ","
	}
	m.description = m.description + desc
	if m.requestMatchers == nil {
		m.requestMatchers = []requestMatcherFunc{}
	}
	m.requestMatchers = append(m.requestMatchers, matcher)
}

func (m *requestMatcher) String() string {
	return m.description
}
