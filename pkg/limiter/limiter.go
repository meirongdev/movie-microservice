// Package limiter provides rate limiting functionality.
package limiter

import "golang.org/x/time/rate"

type Limiter struct {
	limiter *rate.Limiter
}

func New(limit int, burst int) *Limiter {
	l := rate.NewLimiter(rate.Limit(limit), burst)
	return &Limiter{limiter: l}
}

func (l *Limiter) Limit() bool {
	return !l.limiter.Allow()
}
