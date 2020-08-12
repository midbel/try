package try

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var (
	ErrAttempt = errors.New("try: max attempt reached")
	ErrAbort   = errors.New("try: abort")
	ErrNoErr   = errors.New("try: no error")
)

type TryFunc func(int) error

func Try(max int, try TryFunc) error {
	if try == nil {
		return nil
	}
	var (
		wait  = time.Second
		limit = time.Second * 64
		curr  = 1
	)
	for curr < max {
		err := try(curr)
		if err == nil || errors.Is(err, ErrNoErr) {
			break
		}
		if errors.Is(err, ErrAbort) {
			return err
		}
		curr++
		time.Sleep(wait)
		if wait < limit {
			wait = time.Duration(1<<curr) * time.Second
			wait += jitter()
		}
	}
	if curr >= max {
		return ErrAttempt
	}
	return nil
}

func jitter() time.Duration {
	n := rand.Intn(1000)
	return time.Duration(n) * time.Millisecond
}
