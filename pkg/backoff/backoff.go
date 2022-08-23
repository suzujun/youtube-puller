package backoff

import (
	"math"
	"sync/atomic"
	"time"
)

// Backoff is backoff policy for retrying an operation.
type Backoff interface {
	Continue() bool
	Wait() <-chan time.Time
	Reset()
}

type fixedInterval struct {
	interval               time.Duration
	numRetries, maxRetries int
}

func NewFixedIntervalBackoff(interval time.Duration, max int) Backoff {
	return &fixedInterval{
		interval:   interval,
		maxRetries: max,
	}
}

func (f *fixedInterval) Wait() <-chan time.Time {
	defer func() {
		f.numRetries++
	}()
	if f.interval == 0 {
		c := make(chan time.Time, 1)
		c <- time.Now()
		return c
	}
	return time.NewTimer(f.interval).C
}

func (f *fixedInterval) Continue() bool {
	return f.numRetries <= f.maxRetries
}

func (f *fixedInterval) Reset() {
	f.numRetries = 0
}

type exponent struct {
	numRetries, maxRetries int32
}

// NewExponentialBackoff returns backoff policy with exponential algorithm.
func NewExponentialBackoff(max int) Backoff {
	return &exponent{
		maxRetries: int32(max),
	}
}

func (e *exponent) Wait() <-chan time.Time {
	defer atomic.AddInt32(&e.numRetries, 1)
	n := atomic.LoadInt32(&e.numRetries)
	wait := time.Duration(math.Exp2(float64(n))) * time.Second
	return time.NewTimer(wait).C
}

func (e *exponent) Continue() bool {
	return atomic.LoadInt32(&e.numRetries) <= e.maxRetries
}

func (e *exponent) Reset() {
	atomic.StoreInt32(&e.numRetries, 0)
}
