package common

import (
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// ErrBodyNotReplayable is returned when RetryDoer wants to retry a request
// but cannot, because the request has a body and no GetBody to replay it.
// Wrapping (errors.Is) lets callers distinguish "we gave up retrying" from
// "the server actually returned this status".
var ErrBodyNotReplayable = errors.New("proof-sdk-go: cannot retry request with non-replayable body")

// HTTPDoer is the minimal interface satisfied by *http.Client and any of the
// SDK middleware doers. It matches oapi-codegen's HttpRequestDoer so any of
// these doers can be passed via WithHTTPClient on the generated clients.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetryConfig controls RetryDoer behavior. Zero values pick sensible
// defaults; see DefaultRetryConfig.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts after the initial
	// request. 0 disables retries.
	MaxRetries int
	// BaseDelay is the starting backoff. Each attempt doubles up to MaxDelay.
	BaseDelay time.Duration
	// MaxDelay caps the exponential backoff.
	MaxDelay time.Duration
	// RetryableStatuses is the set of response status codes that trigger a
	// retry. Defaults to 408, 429, 500, 502, 503, 504.
	RetryableStatuses map[int]bool
	// RetryOnNetworkError, when true, retries when the transport returns an
	// error (connection refused, EOF, timeout, etc.).
	RetryOnNetworkError bool
}

// DefaultRetryConfig returns a config suitable for most callers: 3 retries,
// 200ms base, 5s cap, retry on 408/429/5xx and on network errors.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  200 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		RetryableStatuses: map[int]bool{
			408: true, 429: true,
			500: true, 502: true, 503: true, 504: true,
		},
		RetryOnNetworkError: true,
	}
}

// RetryDoer wraps an HTTPDoer with exponential-backoff retries. It honors
// Retry-After (seconds form) on 429/503 responses, and respects request
// context cancellation while sleeping between attempts.
//
// RetryDoer requires that the request body be replayable. http.NewRequest /
// http.NewRequestWithContext sets req.GetBody automatically for the common
// body types (bytes.Reader, bytes.Buffer, strings.Reader), which oapi-codegen
// uses. Requests without a replayable body are not retried.
type RetryDoer struct {
	inner HTTPDoer
	cfg   RetryConfig
}

// NewRetryDoer wraps inner with retry semantics.
func NewRetryDoer(inner HTTPDoer, cfg RetryConfig) *RetryDoer {
	if cfg.RetryableStatuses == nil {
		cfg.RetryableStatuses = DefaultRetryConfig().RetryableStatuses
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = DefaultRetryConfig().BaseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = DefaultRetryConfig().MaxDelay
	}
	return &RetryDoer{inner: inner, cfg: cfg}
}

// Do executes req with retries.
func (r *RetryDoer) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		// Replay body on retries (attempt > 0).
		if attempt > 0 && req.Body != nil {
			if req.GetBody == nil {
				// Caller asked for retries but supplied a one-shot body.
				// Surface this distinctly so it isn't confused with a real
				// terminal response.
				return resp, ErrBodyNotReplayable
			}
			body, gerr := req.GetBody()
			if gerr != nil {
				return resp, gerr
			}
			req.Body = body
		}

		resp, err = r.inner.Do(req)

		if !r.shouldRetry(attempt, req, resp, err) {
			return resp, err
		}

		// Drain the previous response body before retrying so the connection
		// can be reused.
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		wait := r.backoff(attempt, resp)
		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
}

func (r *RetryDoer) shouldRetry(attempt int, req *http.Request, resp *http.Response, err error) bool {
	if attempt >= r.cfg.MaxRetries {
		return false
	}
	// Don't retry if the caller's context is already done — the transport
	// error in that case is just ctx.Err() bubbling up, not a transient
	// network failure worth backing off for.
	if ctxErr := req.Context().Err(); ctxErr != nil {
		return false
	}
	if err != nil {
		return r.cfg.RetryOnNetworkError
	}
	if resp == nil {
		return false
	}
	return r.cfg.RetryableStatuses[resp.StatusCode]
}

func (r *RetryDoer) backoff(attempt int, resp *http.Response) time.Duration {
	// Honor Retry-After (seconds form) if present. Server hint takes
	// precedence over our local schedule and is not jittered — the server
	// asked for a specific delay.
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
				d := time.Duration(secs) * time.Second
				if d > r.cfg.MaxDelay {
					return r.cfg.MaxDelay
				}
				return d
			}
		}
	}
	// Exponential: base << attempt, capped at MaxDelay. Integer shift avoids
	// the float-overflow footgun of math.Pow for large attempt counts.
	// Cap the shift at 30 so BaseDelay << shift cannot overflow int64 even
	// for second-scale BaseDelay values.
	shift := min(attempt, 30)
	d := r.cfg.BaseDelay << shift
	if d <= 0 || d > r.cfg.MaxDelay {
		d = r.cfg.MaxDelay
	}
	// Full jitter: pick a random duration in [0, d). Spreads concurrent
	// retriers and avoids thundering-herd on shared backends.
	if d > 0 {
		d = time.Duration(rand.Int64N(int64(d) + 1))
	}
	return d
}
