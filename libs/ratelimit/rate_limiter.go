package ratelimit

import (
	"io"
	"time"
)

// writeResponse and writeRequest represent an individual request for
// a Write operation on our rate limited writer.
type writeResponse struct {
	size int
	err  error
}

type writeRequest struct {
	bytes []byte
	rv    chan writeResponse
}

// This is the actual rate limited writer.  Notice everything's
// private.  We don't need to expose anything at all about this
// externally.  It's just an io.WriteCloser as far anyone's concerned.
type rateLimiter struct {
	input       chan writeRequest
	current     []byte
	currentch   chan writeResponse
	currentsent int
	ticker      *time.Ticker
	limit       int
	remaining   int
	output      io.Writer
	quit        chan bool
}

// gcd is used to compute the smallest time ticker possible for
// getting the most accurate rate limit.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// min keeps us from creating too small a ticker (in which case we're
// spending more time processing ticks than moving bytes)
func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Construct a rate limited writer that will pass the given number of
// bytes per second to the provided io.Writer.
//
// Obviously, closing this Writer will not close the underlying
// writer.
func NewRateLimiter(w io.Writer, bps int) io.WriteCloser {
	unit := time.Second
	g := min(1000, gcd(bps, int(unit)))
	unit /= time.Duration(g)
	bps /= g
	rv := &rateLimiter{
		input:     make(chan writeRequest),
		ticker:    time.NewTicker(unit),
		limit:     bps,
		remaining: bps,
		output:    w,
		quit:      make(chan bool),
	}
	go rv.run()
	return rv
}

func (rl *rateLimiter) run() {
	defer rl.ticker.Stop()
	for {
		// If there's input being processed, don't get
		// anymore.
		//
		// By having the input channel be conditional, it's
		// simply ignored by the select loop when there's
		// already input being processed (in which case the
		// select will only process ticks to write and close
		// to exit).
		input := rl.input
		if rl.current != nil {
			input = nil
		}
		select {
		case <-rl.quit:
			// There's a pending writer.  Tell it we're closed.
			if rl.current != nil {
				rl.currentch <- writeResponse{rl.currentsent, io.EOF}
			}
			return
		case <-rl.ticker.C:
			rl.remaining = rl.limit
			rl.sendSome()
		case req := <-input:
			rl.current = req.bytes
			rl.currentch = req.rv
			rl.currentsent = 0
			rl.sendSome()
		}
	}
}

// Send as many bytes as we can in the current window.
func (rl *rateLimiter) sendSome() {
	if rl.current != nil && rl.remaining > 0 {
		tosend := rl.current
		if rl.remaining < len(rl.current) {
			tosend = rl.current[:rl.remaining]
		}
		rl.remaining -= len(tosend)
		sent, err := rl.output.Write(tosend)
		rl.currentsent += sent
		rl.current = rl.current[sent:]

		if len(rl.current) == 0 || err != nil {
			rl.current = nil
			rl.currentch <- writeResponse{rl.currentsent, err}
		}
	}
}

// io.Writer implementation
func (rl *rateLimiter) Write(p []byte) (n int, err error) {
	req := writeRequest{p, make(chan writeResponse, 1)}
	select {
	case <-rl.quit:
		return 0, io.EOF
	case rl.input <- req:
		res := <-req.rv
		return res.size, res.err
	}
	panic("unreachable")
}

// io.Closer implementation
func (rl *rateLimiter) Close() error {
	// This protects against a double sequential close, but not a
	// double concurrent close.  The latter will panic.
	select {
	case <-rl.quit:
		return io.EOF
	default:
	}
	close(rl.quit)
	return nil
}
