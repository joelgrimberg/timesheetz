package dbcheck

import (
	"strings"
	"testing"
	"time"
)

func TestPingPostgresURL_Unreachable(t *testing.T) {
	// Port 1 is reserved and never listens — connection must fail fast.
	start := time.Now()
	_, err := PingPostgresURL("postgres://x:y@127.0.0.1:1/d?sslmode=disable&connect_timeout=2")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
	if elapsed := time.Since(start); elapsed > 8*time.Second {
		t.Fatalf("ping took too long to fail: %v", elapsed)
	}
}

func TestPingPostgresURL_MalformedURL(t *testing.T) {
	_, err := PingPostgresURL("not-a-url")
	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
	if !strings.Contains(err.Error(), "") { // any error is fine; presence is what we assert
		t.Fatalf("unexpected error shape: %v", err)
	}
}
