package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"proxy-node/internal/core"
)

func TestDefaultProbeURL(t *testing.T) {
	t.Parallel()

	if got, want := defaultProbeURL, "https://www.cloudflare.com/cdn-cgi/trace"; got != want {
		t.Fatalf("defaultProbeURL = %q, want %q", got, want)
	}
}

func TestDefaultSpeedURL(t *testing.T) {
	t.Parallel()

	if got, want := defaultSpeedURL, "https://speed.cloudflare.com/__down?bytes=10000000"; got != want {
		t.Fatalf("defaultSpeedURL = %q, want %q", got, want)
	}
}

func TestDefaultSpeedRetries(t *testing.T) {
	t.Parallel()

	if got, want := defaultSpeedRetries, 1; got != want {
		t.Fatalf("defaultSpeedRetries = %d, want %d", got, want)
	}
}

func TestUsage_ContainsDefaultProbeURL(t *testing.T) {
	t.Parallel()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	usage()
	_ = w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	text := string(out)
	if !strings.Contains(text, defaultProbeURL) {
		t.Fatalf("usage output does not contain default probe URL %q", defaultProbeURL)
	}
	if !strings.Contains(text, defaultSpeedURL) {
		t.Fatalf("usage output does not contain default speed URL %q", defaultSpeedURL)
	}
	if !strings.Contains(text, "--retries int") {
		t.Fatalf("usage output does not contain speed retries flag")
	}
}

func TestCoreLogTails_ContainsBothLogs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	errorPath := filepath.Join(dir, "core.log")
	accessPath := filepath.Join(dir, "access.log")
	if err := os.WriteFile(errorPath, []byte("err-line"), 0o600); err != nil {
		t.Fatalf("WriteFile(error log) error = %v", err)
	}
	if err := os.WriteFile(accessPath, []byte("acc-line"), 0o600); err != nil {
		t.Fatalf("WriteFile(access log) error = %v", err)
	}

	text := coreLogTails(&core.Started{
		LogPath:       errorPath,
		AccessLogPath: accessPath,
	})
	if !strings.Contains(text, "err-line") {
		t.Fatalf("coreLogTails() missing error log content: %q", text)
	}
	if !strings.Contains(text, "acc-line") {
		t.Fatalf("coreLogTails() missing access log content: %q", text)
	}
}

func TestSpeedHTTPWithRetries_SucceedsAfterRetry(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	n, elapsed, attempt, partialErr, err := speedHTTPWithRetries(context.Background(), 3, func(_ context.Context) (int64, time.Duration, error) {
		c := calls.Add(1)
		if c < 2 {
			return 0, 0, context.DeadlineExceeded
		}
		return 2048, 2 * time.Second, nil
	})

	if err != nil {
		t.Fatalf("speedHTTPWithRetries() err = %v, want nil", err)
	}
	if partialErr != nil {
		t.Fatalf("speedHTTPWithRetries() partialErr = %v, want nil", partialErr)
	}
	if attempt != 2 {
		t.Fatalf("attempt = %d, want 2", attempt)
	}
	if n != 2048 {
		t.Fatalf("bytes = %d, want 2048", n)
	}
	if elapsed != 2*time.Second {
		t.Fatalf("elapsed = %v, want 2s", elapsed)
	}
}

func TestSpeedHTTPWithRetries_ReturnsPartialOnReadError(t *testing.T) {
	t.Parallel()

	n, elapsed, attempt, partialErr, err := speedHTTPWithRetries(context.Background(), 3, func(_ context.Context) (int64, time.Duration, error) {
		return 512, 200 * time.Millisecond, io.ErrUnexpectedEOF
	})

	if err != nil {
		t.Fatalf("speedHTTPWithRetries() err = %v, want nil", err)
	}
	if partialErr == nil {
		t.Fatal("partialErr = nil, want non-nil")
	}
	if attempt != 1 {
		t.Fatalf("attempt = %d, want 1", attempt)
	}
	if n != 512 {
		t.Fatalf("bytes = %d, want 512", n)
	}
	if elapsed != 200*time.Millisecond {
		t.Fatalf("elapsed = %v, want 200ms", elapsed)
	}
}
