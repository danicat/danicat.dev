package main

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
)

const (
	defaultMaxSessions = 256
	defaultSessionTTL  = 30 * time.Minute
)

type sessionBucket[State any] struct {
	key        string
	lastAccess time.Time
	store      *localstore.InMemorySessionStore[State]
	snapshots  map[string]struct{}
}

// boundedSessionStore wraps localstore.InMemorySessionStore with a hard LRU
// session limit and idle TTL so unauthenticated traffic cannot exhaust container RAM.
type boundedSessionStore[State any] struct {
	mu          sync.Mutex
	maxSessions int
	ttl         time.Duration
	now         func() time.Time
	order       *list.List
	sessions    map[string]*list.Element
	bySnap      map[string]string
}

var (
	_ aix.SessionStore[any]           = (*boundedSessionStore[any])(nil)
	_ aix.SnapshotMetadataReader[any] = (*boundedSessionStore[any])(nil)
)

func newBoundedSessionStore[State any](maxSessions int, ttl time.Duration) *boundedSessionStore[State] {
	return newBoundedSessionStoreWithClock[State](maxSessions, ttl, time.Now)
}

func newBoundedSessionStoreWithClock[State any](maxSessions int, ttl time.Duration, now func() time.Time) *boundedSessionStore[State] {
	if maxSessions <= 0 {
		maxSessions = defaultMaxSessions
	}
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	if now == nil {
		now = time.Now
	}
	return &boundedSessionStore[State]{
		maxSessions: maxSessions,
		ttl:         ttl,
		now:         now,
		order:       list.New(),
		sessions:    make(map[string]*list.Element, maxSessions),
		bySnap:      make(map[string]string),
	}
}

func (s *boundedSessionStore[State]) evictExpiredLocked(now time.Time) {
	for elem := s.order.Back(); elem != nil; {
		prev := elem.Prev()
		b := elem.Value.(*sessionBucket[State])
		if now.Sub(b.lastAccess) <= s.ttl {
			break
		}
		s.removeElementLocked(elem)
		elem = prev
	}
}

func (s *boundedSessionStore[State]) removeElementLocked(elem *list.Element) {
	b := elem.Value.(*sessionBucket[State])
	for snapID := range b.snapshots {
		delete(s.bySnap, snapID)
	}
	delete(s.sessions, b.key)
	s.order.Remove(elem)
}

func (s *boundedSessionStore[State]) lookupSessionLocked(key string, now time.Time) *sessionBucket[State] {
	s.evictExpiredLocked(now)
	elem, ok := s.sessions[key]
	if !ok {
		return nil
	}
	b := elem.Value.(*sessionBucket[State])
	b.lastAccess = now
	s.order.MoveToFront(elem)
	return b
}

func (s *boundedSessionStore[State]) GetSnapshot(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[State], error) {
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshotID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.evictExpiredLocked(now)
	key, ok := s.bySnap[snapshotID]
	if !ok {
		return nil, nil
	}
	b := s.lookupSessionLocked(key, now)
	if b == nil {
		return nil, nil
	}
	return b.store.GetSnapshot(ctx, snapshotID)
}

func (s *boundedSessionStore[State]) GetSnapshotMetadata(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[State], error) {
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshotID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.evictExpiredLocked(now)
	key, ok := s.bySnap[snapshotID]
	if !ok {
		return nil, nil
	}
	b := s.lookupSessionLocked(key, now)
	if b == nil {
		return nil, nil
	}
	return b.store.GetSnapshotMetadata(ctx, snapshotID)
}

func (s *boundedSessionStore[State]) GetLatestSnapshot(ctx context.Context, sessionID string) (*aix.SessionSnapshot[State], error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	b := s.lookupSessionLocked(sessionID, s.now())
	if b == nil {
		return nil, nil
	}
	return b.store.GetLatestSnapshot(ctx, sessionID)
}

func (s *boundedSessionStore[State]) GetLatestSnapshotMetadata(ctx context.Context, sessionID string) (*aix.SessionSnapshot[State], error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	b := s.lookupSessionLocked(sessionID, s.now())
	if b == nil {
		return nil, nil
	}
	return b.store.GetLatestSnapshotMetadata(ctx, sessionID)
}

func (s *boundedSessionStore[State]) SaveSnapshot(
	ctx context.Context,
	snapshotID string,
	fn func(existing *aix.SessionSnapshot[State]) (*aix.SessionSnapshot[State], error),
) (*aix.SessionSnapshot[State], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.evictExpiredLocked(now)

	// Existing snapshot rewrite: route directly to its owning bucket.
	if snapshotID != "" {
		if existingKey, ok := s.bySnap[snapshotID]; ok {
			if b := s.lookupSessionLocked(existingKey, now); b != nil {
				return b.store.SaveSnapshot(ctx, snapshotID, fn)
			}
		}
	}

	// Evaluate fn on a temporary store first to discover the snapshot's SessionID
	// without allocating or evicting a bucket if fn declines to write or returns an error.
	probe := localstore.NewInMemorySessionStore[State]()
	saved, err := probe.SaveSnapshot(ctx, snapshotID, fn)
	if err != nil || saved == nil {
		return saved, err
	}

	key := saved.SessionID
	if key == "" {
		key = "__anon__:" + saved.SnapshotID
	}

	if b := s.lookupSessionLocked(key, now); b != nil {
		persisted, err := b.store.SaveSnapshot(ctx, saved.SnapshotID, func(_ *aix.SessionSnapshot[State]) (*aix.SessionSnapshot[State], error) {
			return saved, nil
		})
		if err != nil || persisted == nil {
			return persisted, err
		}
		b.snapshots[persisted.SnapshotID] = struct{}{}
		s.bySnap[persisted.SnapshotID] = key
		return persisted, nil
	}

	// Brand new session bucket: reuse probe directly so we don't clone twice.
	for s.order.Len() >= s.maxSessions {
		if tail := s.order.Back(); tail != nil {
			s.removeElementLocked(tail)
		}
	}
	b := &sessionBucket[State]{
		key:        key,
		lastAccess: now,
		store:      probe,
		snapshots:  map[string]struct{}{saved.SnapshotID: {}},
	}
	elem := s.order.PushFront(b)
	s.sessions[key] = elem
	s.bySnap[saved.SnapshotID] = key
	return saved, nil
}
