package main

import (
	"context"
	"testing"
	"time"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/sessionstoretest"
)

func TestBoundedSessionStoreConformance(t *testing.T) {
	sessionstoretest.Run(t, func(t *testing.T) aix.SessionStore[any] {
		return newBoundedSessionStore[any](64, time.Hour)
	}, nil)
}

func TestBoundedSessionStoreLRUAndTTL(t *testing.T) {
	t.Parallel()

	current := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	store := newBoundedSessionStoreWithClock[any](2, 10*time.Minute, func() time.Time { return current })
	ctx := context.Background()

	save := func(sessionID string) string {
		t.Helper()
		snap, err := store.SaveSnapshot(ctx, "", func(_ *aix.SessionSnapshot[any]) (*aix.SessionSnapshot[any], error) {
			return &aix.SessionSnapshot[any]{
				SessionID: sessionID,
				CreatedAt: current,
				UpdatedAt: current,
				Status:    aix.SnapshotStatusCompleted,
			}, nil
		})
		if err != nil || snap == nil {
			t.Fatalf("SaveSnapshot(%q) failed: snap=%v err=%v", sessionID, snap, err)
		}
		return snap.SnapshotID
	}

	snapA := save("sess-a")
	current = current.Add(1 * time.Minute)
	snapB := save("sess-b")

	// Touch sess-a so sess-b becomes least recently used.
	current = current.Add(1 * time.Minute)
	if meta, err := store.GetLatestSnapshotMetadata(ctx, "sess-a"); err != nil || meta == nil {
		t.Fatalf("expected sess-a to exist: meta=%v err=%v", meta, err)
	}

	// Saving sess-c must evict sess-b (LRU) while keeping sess-a and sess-c.
	current = current.Add(1 * time.Minute)
	save("sess-c")

	if meta, err := store.GetLatestSnapshotMetadata(ctx, "sess-b"); err != nil || meta != nil {
		t.Fatalf("expected evicted sess-b to return nil, got meta=%v err=%v", meta, err)
	}
	if snap, err := store.GetSnapshot(ctx, snapB); err != nil || snap != nil {
		t.Fatalf("expected evicted snapshot %q to return nil, got snap=%v err=%v", snapB, snap, err)
	}
	if snap, err := store.GetSnapshot(ctx, snapA); err != nil || snap == nil {
		t.Fatalf("expected sess-a snapshot %q to remain, got snap=%v err=%v", snapA, snap, err)
	}

	// Advance clock past TTL -> both remaining sessions expire.
	current = current.Add(11 * time.Minute)
	for _, sid := range []string{"sess-a", "sess-c"} {
		if meta, err := store.GetLatestSnapshotMetadata(ctx, sid); err != nil || meta != nil {
			t.Fatalf("expected expired %q to return nil, got meta=%v err=%v", sid, meta, err)
		}
	}
}
