package limiter

import "golang.org/x/time/rate"

type Limiter struct {
	l *rate.Limiter
}

func New(limit int, burst int) *Limiter {
	return &Limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *Limiter) Limit() bool {
	return l.l.Allow()
}
