package retry

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcCall_Success(t *testing.T) {
	err := GrpcCall(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGrpcCall_RetryableError(t *testing.T) {
	attempts := 0
	err := GrpcCall(func() error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "service unavailable")
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestGrpcCall_NonRetryableError(t *testing.T) {
	err := GrpcCall(func() error {
		return status.Error(codes.InvalidArgument, "invalid argument")
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestGrpcCall_MaxRetriesExceeded(t *testing.T) {
	start := time.Now()
	err := GrpcCall(func() error {
		return status.Error(codes.Unavailable, "service unavailable")
	})
	elapsed := time.Since(start)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if elapsed < 300*time.Millisecond {
		t.Errorf("expected at least 300ms of retries, got %v", elapsed)
	}
}

func TestGrpcCall_NonGrpcError(t *testing.T) {
	err := GrpcCall(func() error {
		return errors.New("some non-grpc error")
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
