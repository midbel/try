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
	DefaultDelay   = time.Second
	DefaultBackoff = DefaultDelay * 64
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

func Foreover(try TryFunc) error {
	r, _ := New(0)
	return r.Try(try)
}

func Try(max int, try TryFunc) error {
	r, _ := New(max)
	return r.Try(try)
}

func New(limit int, options ...Option) (*Retry, error) {
	r := Retry{
		limit:   limit,
		wait:    DefaultDelay,
		backoff: DefaultBackoff,
	}
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
			return errors.Unwrap(err)
		}
		curr++
		time.Sleep(wait)
		if wait < r.backoff {
			wait = time.Duration(1<<curr) * r.wait
			if r.jitter != nil {
				wait += r.jitter()
			}
		}
	}
	if r.limit > 0 && curr >= r.limit {
		return ErrAttempt
	}
	return nil
}

func jitter() time.Duration {
	n := rand.Intn(1000)
	return time.Duration(n) * time.Millisecond
}
