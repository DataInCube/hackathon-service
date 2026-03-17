package events

import (
	"context"
	"testing"
)

func TestNewNatsPublisher_RequiresURL(t *testing.T) {
	pub, err := NewNatsPublisher("", "", nil)
	if err == nil {
		t.Fatal("expected error for empty nats url")
	}
	if pub != nil {
		t.Fatal("publisher must be nil on error")
	}
}

func TestNatsPublisher_NilSafeMethods(t *testing.T) {
	var pub *NatsPublisher
	if err := pub.Publish(context.Background(), "subject", map[string]any{"ok": true}); err != nil {
		t.Fatalf("nil publisher publish should be no-op, got %v", err)
	}
	pub.Close()
}
