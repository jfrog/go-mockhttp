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

func (m *requestMatcher) Method(method string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Method(%s)", method), func(request *http.Request) bool {
		return request.Method == method
	})
	return m
}

func (m *requestMatcher) Path(path string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Path(%s)", path), func(request *http.Request) bool {
		return request.URL.Path == path
	})
	return m
}

func (m *requestMatcher) PathMatches(path *regexp.Regexp) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("PathMatches(%s)", path), func(request *http.Request) bool {
		return path.MatchString(request.URL.Path)
	})
	return m
}

func (m *requestMatcher) GET(path string) *requestMatcher {
	return m.Method("GET").Path(path)
}

func (m *requestMatcher) POST(path string) *requestMatcher {
	return m.Method("POST").Path(path)
}

func (m *requestMatcher) Header(key string, value string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("Header(%s: %s)", key, value), func(request *http.Request) bool {
		return request.Header.Get(key) == value
	})
	return m
}

func (m *requestMatcher) HeaderMatches(key string, value *regexp.Regexp) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("HeaderMatches(%s: %s)", key, value), func(request *http.Request) bool {
		return value.MatchString(request.Header.Get(key))
	})
	return m
}

func (m *requestMatcher) NoHeader(key string) *requestMatcher {
	m.appendMatcher(fmt.Sprintf("NoHeader(%s)", key), func(request *http.Request) bool {
		return request.Header.Get(key) == ""
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
