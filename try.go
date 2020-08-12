package try

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	DefaultLimitMax   = 10
	DefaultDelay      = time.Second
	DefaultBackoffMax = DefaultDelay * 64
)

var (
	ErrAttempt  = errors.New("try: max attempt reached")
	ErrAbort    = errors.New("try: abort")
	ErrNoErr    = errors.New("try: no error")
	ErrDuration = errors.New("try: invalid duration")
)

type (
	TryFunc    func(int) error
	JitterFunc func() time.Duration
)

type Option func(r *Retry) error

func WithWait(d time.Duration) Option {
	return func(r *Retry) error {
		if d <= 0 {
			return ErrDuration
		}
		r.wait = d
		return nil
	}
}

func WithBackoff(d time.Duration) Option {
	return func(r *Retry) error {
		if d <= 0 {
			return ErrDuration
		}
		r.backoff = d
		return nil
	}
}

func WithJitter(fn JitterFunc) Option {
	return func(r *Retry) error {
		if fn == nil {
			fn = jitter
		}
		r.jitter = fn
		return nil
	}
}

type Retry struct {
	limit   int
	wait    time.Duration
	backoff time.Duration
	jitter  JitterFunc
}

func Try(max int, try TryFunc) error {
	r, _ := New(max)
	return r.Try(try)
}

func New(limit int, options ...Option) (*Retry, error) {
	var r Retry
	r.init()
	for _, o := range options {
		if err := o(&r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

func (r *Retry) Try(try TryFunc) error {
	if try == nil {
		return nil
	}
	var (
		wait = r.wait
		curr int
	)
	for r.limit == 0 || curr < r.limit {
		err := try(curr)
		if err == nil || errors.Is(err, ErrNoErr) {
			break
		}
		if errors.Is(err, ErrAbort) {
			return err
		}
		curr++
		time.Sleep(wait)
		if curr > 1 && wait < r.backoff {
			wait = time.Duration(1<<curr) * r.wait
			wait += jitter()
		}
	}
	if curr >= r.limit {
		return ErrAttempt
	}
	return nil
}

func (r *Retry) init() {
	r.limit = DefaultLimitMax
	r.wait = DefaultDelay
	r.backoff = DefaultBackoffMax
	r.jitter = jitter
}

func jitter() time.Duration {
	n := rand.Intn(1000)
	return time.Duration(n) * time.Millisecond
}
