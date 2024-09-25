package retry

import (
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcWhiteListClassifier struct {
	whitelist map[codes.Code]struct{}
}

func (g grpcWhiteListClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}
	e, ok := status.FromError(err)
	if !ok {
		return retrier.Fail
	}
	if _, ok := g.whitelist[e.Code()]; ok {
		return retrier.Retry
	}
	return retrier.Fail
}

func grpcRetry(fn func() error, whitelist map[codes.Code]struct{}) error {
	r := retrier.New(retrier.ExponentialBackoff(3, time.Millisecond*100), &grpcWhiteListClassifier{whitelist})
	r.SetJitter(0.25)
	err := r.Run(func() error {
		return fn()
	})
	return err
}

func GrpcCall(fn func() error) error {
	whitelist := map[codes.Code]struct{}{
		codes.Unavailable:       {},
		codes.DeadlineExceeded:  {},
		codes.ResourceExhausted: {},
	}
	return grpcRetry(fn, whitelist)
}
