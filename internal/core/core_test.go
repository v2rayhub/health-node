package core

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerStart_WritesFinalLogConfig(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r := Runner{
		CorePath: "/bin/true",
		Port:     1080,
		Timeout:  5 * time.Second,
	}

	outbound := map[string]any{
		"tag":      "proxy",
		"protocol": "freedom",
	}

	started, err := r.Start(ctx, outbound)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer started.Stop()

	raw, err := os.ReadFile(started.ConfigPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", started.ConfigPath, err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("Unmarshal(config) error = %v", err)
	}

	logCfg, ok := cfg["log"].(map[string]any)
	if !ok {
		t.Fatalf("config.log missing or wrong type: %#v", cfg["log"])
	}

	if got := logCfg["access"]; got != started.AccessLogPath {
		t.Fatalf("log.access = %#v, want %q", got, started.AccessLogPath)
	}
	if got := logCfg["error"]; got != started.LogPath {
		t.Fatalf("log.error = %#v, want %q", got, started.LogPath)
	}

	if filepath.Dir(started.ConfigPath) == "" {
		t.Fatalf("config dir is empty for %q", started.ConfigPath)
	}
}

func TestStarted_ReadAccessLogTail(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	accessPath := filepath.Join(dir, "access.log")
	content := strings.Repeat("x", 4500)
	if err := os.WriteFile(accessPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", accessPath, err)
	}

	s := &Started{AccessLogPath: accessPath}
	got := s.ReadAccessLogTail()
	if len(got) != 4000 {
		t.Fatalf("len(ReadAccessLogTail()) = %d, want 4000", len(got))
	}
}
