// An io.Writer to rate limit writes to other io.Writers.

package main

import (
	. "GoRedis/libs/ratelimit"
	"fmt"
	"io"
	"math/rand"
	"os"
	"unicode"
	"unicode/utf8"
)

// This is just a creepy text effect over a writer just to demonstrate
// more io layering.

type caperturber struct {
	io.Writer
}

func (c *caperturber) Write(b []byte) (written int, err error) {
	for _, r := range string(b) {
		if rand.Intn(4) == 0 {
			r = unicode.ToUpper(r)
		}
		buf := []byte{0, 0, 0, 0}
		n := utf8.EncodeRune(buf, r)
		buf = buf[:n]
		w, err := c.Writer.Write(buf)
		written += w
		if err != nil {
			break
		}
	}
	return
}

func main() {
	// 10 bytes per second over stdout.
	rl := NewRateLimiter(os.Stdout, 10)
	defer rl.Close()

	// Here we see how we can combine io.Writers.  In this case,
	// we've got a capitalization filter manipulating text before
	// the rate limiter slows it down on its way to stdout.
	w := &caperturber{rl}

	// And you use it like any other io.Writer
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "all work and no play makes jack a dull boy\n")
	}
}
