package union

import (
	"context"
	"net/http"
	"time"
)

// MaxRetry is the maximum number of retries before stopping.
var MaxRetry = 10

// MaxJitter will randomize over the full exponential backoff time
const MaxJitter = 1.0

// NoJitter disables the use of jitter for randomizing the exponential backoff time
const NoJitter = 0.0

// DefaultRetryUnit - default unit multiplicative per retry.
// defaults to 200 * time.Millisecond
var DefaultRetryUnit = 200 * time.Millisecond

// DefaultRetryCap - Each retry attempt never waits no longer than
// this maximum time duration.
var DefaultRetryCap = time.Second

// newRetryTimer creates a timer with exponentially increasing
// delays until the maximum retry attempts are reached.
func (c Client) newRetryTimer(ctx context.Context, maxRetry int, unit time.Duration, cap time.Duration, jitter float64) <-chan int {
	attemptCh := make(chan int)

	// computes the exponential backoff duration according to
	// https://www.awsarchitectureblog.com/2015/03/backoff.html
	exponentialBackoffWait := func(attempt int) time.Duration {
		// normalize jitter to the range [0, 1.0]
		if jitter < NoJitter {
			jitter = NoJitter
		}
		if jitter > MaxJitter {
			jitter = MaxJitter
		}

		//sleep = random_between(0, min(cap, base * 2 ** attempt))
		sleep := unit * time.Duration(1<<uint(attempt))
		if sleep > cap {
			sleep = cap
		}
		if jitter != NoJitter {
			sleep -= time.Duration(c.random.Float64() * float64(sleep) * jitter)
		}
		return sleep
	}

	go func() {
		defer close(attemptCh)
		for i := 0; i < maxRetry; i++ {
			select {
			case attemptCh <- i + 1:
			case <-ctx.Done():
				return
			}

			select {
			case <-time.After(exponentialBackoffWait(i)):
			case <-ctx.Done():
				return
			}
		}
	}()
	return attemptCh
}

// List of error codes which are retryable.
var retryableCodes = map[string]struct{}{
	"RequestError":          {},
	"RequestTimeout":        {},
	"Throttling":            {},
	"ThrottlingException":   {},
	"RequestLimitExceeded":  {},
	"RequestThrottled":      {},
	"InternalError":         {},
	"ExpiredToken":          {},
	"ExpiredTokenException": {},
	"SlowDown":              {},
}

// is error code retryable.
func isRetryable(code string) (ok bool) {
	_, ok = retryableCodes[code]
	return ok
}

// List of HTTP status codes which are retryable.
var retryableHTTPStatusCodes = map[int]struct{}{
	429:                            {}, // http.StatusTooManyRequests is not part of the Go 1.5 library, yet
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	http.StatusGatewayTimeout:      {},
	// Add more HTTP status codes here.
}

// isHTTPStatusRetryable - is HTTP error code retryable.
func isHTTPStatusRetryable(httpStatusCode int) (ok bool) {
	_, ok = retryableHTTPStatusCodes[httpStatusCode]
	return ok
}
