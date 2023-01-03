package try

import (
	"context"
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

func Forever(try TryFunc) error {
	return Try(0, try)
}

func TryContext(ctx context.Context, max int, try TryFunc) error {
	r, err := New(max)
	if err != nil {
		return err
	}
	return r.TryContext(ctx, try)
}

func Try(max int, try TryFunc) error {
	return TryContext(context.TODO(), max, try)
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

func (r *Retry) TryContext(ctx context.Context, try TryFunc) error {
	if try == nil {
		return nil
	}
	if r.limit == 1 {
		return try(0)
	}
	var (
		wait = r.wait
		curr int
	)
	for r.limit == 0 || curr < r.limit {
		if err := ctx.Err(); err != nil {
			return err
		}
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

func (r *Retry) Try(try TryFunc) error {
	return r.TryContext(context.TODO(), try)
}

func jitter() time.Duration {
	n := rand.Intn(1000)
	return time.Duration(n) * time.Millisecond
}
