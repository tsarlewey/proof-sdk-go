package common

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDoer returns scripted responses/errors in order. Each entry is
// either a response or a transport error. Calls past the script length
// return the last entry repeatedly. The doer captures each request's body
// bytes (drained inside Do) so tests can assert on body replay.
type stubDoer struct {
	responses  []*http.Response
	errs       []error
	calls      int32
	bodiesSeen [][]byte
	mu         sync.Mutex
}

func (s *stubDoer) Do(req *http.Request) (*http.Response, error) {
	idx := int(atomic.AddInt32(&s.calls, 1) - 1)

	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	s.mu.Lock()
	s.bodiesSeen = append(s.bodiesSeen, body)
	s.mu.Unlock()

	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	return s.responses[idx], s.errs[idx]
}

func TestRetryDoer_NoRetryOnSuccess(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(200, nil)},
		errs:      []error{nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := r.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.EqualValues(t, 1, stub.calls)
}

func TestRetryDoer_RetriesOn503ThenSucceeds(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{
			MockJSONResponse(503, nil),
			MockJSONResponse(503, nil),
			MockJSONResponse(200, nil),
		},
		errs: []error{nil, nil, nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := r.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.EqualValues(t, 3, stub.calls)
}

func TestRetryDoer_GivesUpAfterMaxRetries(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(500, nil)},
		errs:      []error{nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := r.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	// Initial + 2 retries = 3 calls.
	assert.EqualValues(t, 3, stub.calls)
}

func TestRetryDoer_DoesNotRetry4xx(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(404, nil)},
		errs:      []error{nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 5, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	_, err := r.Do(req)
	require.NoError(t, err)
	assert.EqualValues(t, 1, stub.calls)
}

func TestRetryDoer_RetryOnNetworkError(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{nil, MockJSONResponse(200, nil)},
		errs:      []error{errors.New("connection refused"), nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond, RetryOnNetworkError: true})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := r.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.EqualValues(t, 2, stub.calls)
}

func TestRetryDoer_HonorsRetryAfter(t *testing.T) {
	respWithRA := MockJSONResponse(429, nil)
	respWithRA.Header.Set("Retry-After", "1")
	stub := &stubDoer{
		responses: []*http.Response{respWithRA, MockJSONResponse(200, nil)},
		errs:      []error{nil, nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 1, BaseDelay: 10 * time.Millisecond, MaxDelay: 2 * time.Second})
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	start := time.Now()
	resp, err := r.Do(req)
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	// Retry-After said 1s; verify we waited close to that, not the 10ms base.
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond)
}

func TestRetryDoer_RespectsContextCancellation(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(503, nil), MockJSONResponse(200, nil)},
		errs:      []error{nil, nil},
	}
	// Long backoff so context cancels before the next attempt.
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 3, BaseDelay: time.Hour, MaxDelay: time.Hour})

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := r.Do(req)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestRetryDoer_ReplaysBodyContent(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(503, nil), MockJSONResponse(200, nil)},
		errs:      []error{nil, nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	const payload = `{"k":"v"}`
	req, _ := http.NewRequest("POST", "https://example.com", strings.NewReader(payload))

	resp, err := r.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	require.Len(t, stub.bodiesSeen, 2)
	assert.Equal(t, payload, string(stub.bodiesSeen[0]))
	assert.Equal(t, payload, string(stub.bodiesSeen[1]), "retry must carry the same body")
}

func TestRetryDoer_NonReplayableBodyReturnsTypedError(t *testing.T) {
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(503, nil), MockJSONResponse(200, nil)},
		errs:      []error{nil, nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})

	// nopCloserBody returns a Body but no GetBody — mimics a request whose
	// body isn't replayable (e.g., a streaming reader).
	req, _ := http.NewRequest("POST", "https://example.com", nil)
	req.Body = io.NopCloser(strings.NewReader("once"))
	req.GetBody = nil

	_, err := r.Do(req)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBodyNotReplayable), "expected ErrBodyNotReplayable, got %v", err)
	// Only the initial call should have happened — no retry attempted.
	assert.EqualValues(t, 1, stub.calls)
}

func TestRetryDoer_BackoffBoundedByMaxDelay(t *testing.T) {
	// Exhaust retries with a tiny MaxDelay to verify bounded sleeps even at
	// high attempt counts (no overflow from large shifts).
	stub := &stubDoer{
		responses: []*http.Response{MockJSONResponse(500, nil)},
		errs:      []error{nil},
	}
	r := NewRetryDoer(stub, RetryConfig{MaxRetries: 5, BaseDelay: time.Second, MaxDelay: 5 * time.Millisecond})
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	start := time.Now()
	_, err := r.Do(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	// 5 retries × max 5ms each = 25ms ceiling; allow generous slack for CI.
	assert.Less(t, elapsed, 200*time.Millisecond, "MaxDelay must cap shifted backoff")
	assert.EqualValues(t, 6, stub.calls)
}
