package mockhttp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func newRequestRecorder() *requestRecorder {
	return &requestRecorder{
		acceptedRequests:  []recordedRequest{},
		unmatchedRequests: []recordedRequest{},
	}
}

type requestRecorder struct {
	acceptedRequests  []recordedRequest
	unmatchedRequests []recordedRequest
}

func (r *requestRecorder) AcceptedRequests() []recordedRequest {
	return r.acceptedRequests
}

func (r *requestRecorder) recordAcceptedRequest(req *http.Request) {
	r.acceptedRequests = append(r.acceptedRequests, newRecordedRequest(req))
}

func (r *requestRecorder) UnmatchedRequests() []recordedRequest {
	return r.unmatchedRequests
}

func (r *requestRecorder) recordUnmatchedRequest(req *http.Request) {
	r.unmatchedRequests = append(r.unmatchedRequests, newRecordedRequest(req))
}

func (r *requestRecorder) ClearHistory() {
	r.acceptedRequests = []recordedRequest{}
	r.unmatchedRequests = []recordedRequest{}
}

func (r *requestRecorder) String() string {
	requests2str := func(requests []recordedRequest) string {
		str := ""
		for i, req := range requests {
			str += fmt.Sprintf("  %2d: %s %s\n", i+1, req.Method, req.Path)
			for k, v := range req.Header {
				str += fmt.Sprintf("      %s: %s\n", k, v)
			}
		}
		return str
	}
	accepted := requests2str(r.acceptedRequests)
	unmatched := requests2str(r.unmatchedRequests)
	return fmt.Sprintf(""+
		"Acceppted:\n%s"+
		"Unmatched:\n%s", accepted, unmatched)
}

type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

func newRecordedRequest(r *http.Request) recordedRequest {
	return recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header,
		Body:   readAllOrNil(r.Body),
	}
}

func readAllOrNil(r io.Reader) []byte {
	if r == nil {
		return nil
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil
	}
	return data
}

func (r recordedRequest) toHttpRequest() *http.Request {
	httpRequest := http.Request{
		Method: r.Method,
		URL: &url.URL{
			Path:     r.Path,
			RawQuery: r.Query.Encode(),
		},
		Header: r.Header,
		Body:   ioutil.NopCloser(bytes.NewReader(r.Body)),
	}
	return &httpRequest
}

func (r recordedRequest) BodyAsString() string {
	if r.Body != nil {
		return string(r.Body)
	}
	return ""
}

func (r recordedRequest) String() string {
	b := strings.Builder{}
	query := ""
	if len(r.Query) > 0 {
		query = fmt.Sprintf("?%s", r.Query.Encode())
	}
	b.WriteString(fmt.Sprintf("%s %s%s\n", r.Method, r.Path, query))
	r.Header.Write(&b)
	b.WriteString(r.BodyAsString())
	b.WriteRune('\n')
	return b.String()
}
