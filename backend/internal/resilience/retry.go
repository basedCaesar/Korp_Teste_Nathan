package resilience

import (
	"context"
	"time"
)

func Retry(ctx context.Context, tentativas int, backoff time.Duration, retryable func(error) bool, fn func(context.Context) error) error {
	var err error
	for tentativa := 1; tentativa <= tentativas; tentativa++ {
		err = fn(ctx)
		if err == nil {
			return nil
		}
		if tentativa == tentativas || !retryable(err) {
			return err
		}
		select {
		case <-time.After(backoff * time.Duration(tentativa)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
