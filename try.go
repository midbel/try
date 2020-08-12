package try

import (
	"errors"
	"time"
)

var ErrAttempt = errors.New("try: max attempt reached")

type TryFunc func(int) error

func Try(max int, try TryFunc) error {
	if try == nil {
		return nil
	}
	var (
		wait  = time.Second * 5
		limit = time.Second * 30
		curr  int
	)
	for curr < max {
		err := try(curr + 1)
		if err == nil {
			break
		}
		time.Sleep(wait)
		if wait < limit {
			wait += time.Second
		}
		curr++
	}
	if curr >= max {
		return ErrAttempt
	}
	return nil
}
