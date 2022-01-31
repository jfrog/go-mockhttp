package mockhttp

import (
	"fmt"
)

func newVerifier(matcher *requestMatcher, opts ...verifyOpt) *verifier {
	verifier := &verifier{
		matcher: matcher,
	}
	Once()(verifier)
	for _, opt := range opts {
		opt(verifier)
	}
	return verifier
}

type verifier struct {
	matcher     *requestMatcher
	verifyTimes func(acceptedCount, unmatchedCount int) error
}

func (v *verifier) verifyRequests(recorder *requestRecorder) error {
	acceptedCount := v.countRequests(recorder.AcceptedRequests())
	unmatchedCount := v.countRequests(recorder.UnmatchedRequests())
	if err := v.verifyTimes(acceptedCount, unmatchedCount); err != nil {
		return verifyError(fmt.Sprintf("%s\n%s", err, v.detailsStr(recorder)))
	}
	return nil
}

func (v *verifier) countRequests(requests []recordedRequest) int {
	count := 0
	for _, req := range requests {
		if v.matcher.matches(req.toHttpRequest()) {
			count++
		}
	}
	return count
}

func (v *verifier) detailsStr(recorder *requestRecorder) string {
	return fmt.Sprintf(""+
		"expected: %s \n"+
		"actual  : %s", v.matcher, recorder)
}

type verifyOpt func(verifier *verifier)

// Times is a verify functional option to set how many times a request is expected
func Times(expected int) verifyOpt {
	return func(opts *verifier) {
		opts.verifyTimes = func(acceptedCount, unmatchedCount int) error {
			totalCount := acceptedCount + unmatchedCount
			if totalCount != expected {
				return verifyError(fmt.Sprintf("request was called unexpected number of times. expected: %d, actual: %d", expected, totalCount))
			}
			return nil
		}
	}
}

// Once is a verify functional option to set, a request is expected exactly once
func Once() verifyOpt {
	return Times(1)
}

// Never is a verify functional option to set, a request is never expected
func Never() verifyOpt {
	return Times(0)
}

// AtLeast is a verify functional option to set, a request is expected at least number of given times
func AtLeast(times int) verifyOpt {
	return func(opts *verifier) {
		opts.verifyTimes = func(acceptedCount, unmatchedCount int) error {
			totalCount := acceptedCount + unmatchedCount
			if totalCount < times {
				return verifyError(fmt.Sprintf("request was called unexpected number of times. expected at least: %d, actual: %d", times, totalCount))
			}
			return nil
		}
	}
}

// AtMost is a verify functional option to set, a request is expected at most number of given times
func AtMost(times int) verifyOpt {
	return func(opts *verifier) {
		opts.verifyTimes = func(acceptedCount, unmatchedCount int) error {
			totalCount := acceptedCount + unmatchedCount
			if totalCount > times {
				return verifyError(fmt.Sprintf("request was called unexpected number of times. expected at most: %d, actual: %d", times, totalCount))
			}
			return nil
		}
	}
}

type verifyError string

func (e verifyError) Error() string {
	return string(e)
}
