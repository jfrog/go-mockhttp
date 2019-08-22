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

func Once() verifyOpt {
	return Times(1)
}

func Never() verifyOpt {
	return Times(0)
}

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
