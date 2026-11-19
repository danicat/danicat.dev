package main

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	aix "github.com/firebase/genkit/go/ai/exp"
	"golang.org/x/time/rate"
)

const (
	maxBodyBytes      = 8192
	maxPromptRunes    = 1200
	ipRateInterval    = 10 * time.Second // 6 requests/min per IP steady state
	ipRateBurst       = 4
	maxTrackedIPs     = 4096
	globalRateEvery   = 3 * time.Second // 20 requests/min across all IPs combined
	globalRateBurst   = 10
	globalDailyMax    = 1500
	globalDailyWindow = 24 * time.Hour
)

func buildAllowedOrigins(allowLocalhost bool) map[string]struct{} {
	origins := map[string]struct{}{
		"https://danicat.dev":     {},
		"https://www.danicat.dev": {},
	}
	if allowLocalhost {
		origins["http://localhost:1313"] = struct{}{}
		origins["http://127.0.0.1:1313"] = struct{}{}
	}
	return origins
}

type visitorBucket struct {
	ip      string
	limiter *rate.Limiter
}

// ipRateLimiter tracks per-IP token buckets in an O(1) LRU bounded at maxEntries
// so rotating-IP botnets cannot grow memory unboundedly or stall the mutex.
type ipRateLimiter struct {
	mu         sync.Mutex
	maxEntries int
	interval   time.Duration
	burst      int
	order      *list.List
	visitors   map[string]*list.Element
}

func newIPRateLimiter(maxEntries int, interval time.Duration, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		maxEntries: maxEntries,
		interval:   interval,
		burst:      burst,
		order:      list.New(),
		visitors:   make(map[string]*list.Element, min(maxEntries, 256)),
	}
}

func (l *ipRateLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, ok := l.visitors[ip]; ok {
		l.order.MoveToFront(elem)
		return elem.Value.(*visitorBucket).limiter.Allow()
	}
	if l.order.Len() >= l.maxEntries {
		if tail := l.order.Back(); tail != nil {
			evicted := tail.Value.(*visitorBucket)
			delete(l.visitors, evicted.ip)
			l.order.Remove(tail)
		}
	}
	b := &visitorBucket{
		ip:      ip,
		limiter: rate.NewLimiter(rate.Every(l.interval), l.burst),
	}
	l.visitors[ip] = l.order.PushFront(b)
	return b.limiter.Allow()
}

// globalGuard enforces container-wide throughput and rolling 24h request ceilings
// so a distributed rotating-IP botnet cannot bypass per-IP rate limits.
type globalGuard struct {
	mu        sync.Mutex
	limiter   *rate.Limiter
	dailyMax  int
	window    time.Duration
	now       func() time.Time
	ring      []time.Time
	head      int
	count     int
}

func newGlobalGuard(interval time.Duration, burst int, dailyMax int, window time.Duration, now func() time.Time) *globalGuard {
	if now == nil {
		now = time.Now
	}
	return &globalGuard{
		limiter:  rate.NewLimiter(rate.Every(interval), burst),
		dailyMax: dailyMax,
		window:   window,
		now:      now,
		ring:     make([]time.Time, dailyMax),
	}
}

func (g *globalGuard) Allow() (bool, string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.now()
	cutoff := now.Add(-g.window)
	for g.count > 0 && !g.ring[g.head].After(cutoff) {
		g.head = (g.head + 1) % g.dailyMax
		g.count--
	}
	if g.count >= g.dailyMax {
		return false, "global_daily_limit"
	}
	if !g.limiter.AllowN(now, 1) {
		return false, "global_rate_limit"
	}
	tail := (g.head + g.count) % g.dailyMax
	g.ring[tail] = now
	g.count++
	return true, ""
}

func newHTTPMux(
	agentName string,
	sessions aix.SnapshotMetadataReader[any],
	agentHandler http.HandlerFunc,
	allowLocalhost bool,
) *http.ServeMux {
	return newHTTPMuxWithLimiters(
		agentName,
		sessions,
		agentHandler,
		buildAllowedOrigins(allowLocalhost),
		newIPRateLimiter(maxTrackedIPs, ipRateInterval, ipRateBurst),
		newGlobalGuard(globalRateEvery, globalRateBurst, globalDailyMax, globalDailyWindow, time.Now),
	)
}

func newHTTPMuxWithLimiters(
	agentName string,
	sessions aix.SnapshotMetadataReader[any],
	agentHandler http.HandlerFunc,
	origins map[string]struct{},
	ipLimiter *ipRateLimiter,
	global *globalGuard,
) *http.ServeMux {
	cop := http.NewCrossOriginProtection()
	for origin := range origins {
		if err := cop.AddTrustedOrigin(origin); err != nil {
			panic(fmt.Sprintf("invalid trusted origin %q: %v", origin, err))
		}
	}
	agentPath := "/agents/" + agentName

	mux := http.NewServeMux()
	mux.Handle("POST "+agentPath, cop.Handler(withStrictCORS(origins,
		withRateLimits(ipLimiter, global, withStrictAgentInput(sessions, agentHandler)),
	)))
	mux.HandleFunc("OPTIONS "+agentPath, withCORSHeaders(origins, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	return mux
}

func withCORSHeaders(origins map[string]struct{}, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
			if _, ok := origins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
			}
		}
		next(w, r)
	}
}

// withStrictCORS adds CORS headers and rejects POST requests with a missing or
// unlisted Origin header so non-browser callers cannot silently bypass Origin checks.
func withStrictCORS(origins map[string]struct{}, next http.HandlerFunc) http.HandlerFunc {
	return withCORSHeaders(origins, func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if _, ok := origins[origin]; !ok {
			slog.Warn("origin_rejected", "origin", origin, "ip", clientIP(r))
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "forbidden_origin: request origin is not allowed",
			})
			return
		}
		next(w, r)
	})
}

func withRateLimits(ipLimiter *ipRateLimiter, global *globalGuard, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !ipLimiter.Allow(ip) {
			slog.Warn("ip_rate_limited", "ip", ip)
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate_limited: too many requests, please wait a moment",
			})
			return
		}
		if ok, reason := global.Allow(); !ok {
			slog.Warn(reason, "ip", ip)
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate_limited: service is busy, please try again later",
			})
			return
		}
		next(w, r)
	}
}

type sessionInit struct {
	SessionID string `json:"sessionId,omitempty"`
}

type agentEnvelope struct {
	Init *sessionInit `json:"init,omitempty"`
	Data struct {
		Message *struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	} `json:"data"`
}

func validateSession(ctx context.Context, sessions aix.SnapshotMetadataReader[any], init *sessionInit) (int, string) {
	if init == nil || init.SessionID == "" {
		return 0, ""
	}
	snap, err := sessions.GetLatestSnapshotMetadata(ctx, init.SessionID)
	if err != nil {
		return http.StatusInternalServerError, "internal_error: failed to verify session"
	}
	if snap == nil {
		return http.StatusNotFound, "session_not_found: unknown or expired sessionId"
	}
	return 0, ""
}

func withStrictAgentInput(sessions aix.SnapshotMetadataReader[any], next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
				"error": "payload_too_large: request body exceeds 8KB limit",
			})
			return
		}

		var env agentEnvelope
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&env); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid_request: malformed or unsupported JSON fields",
			})
			return
		}
		if status, msg := validateSession(r.Context(), sessions, env.Init); status != 0 {
			writeJSON(w, status, map[string]string{"error": msg})
			return
		}
		msg := env.Data.Message
		if msg == nil || msg.Role != "user" || len(msg.Content) != 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid_request: expected a single user message text part",
			})
			return
		}
		text := strings.TrimSpace(msg.Content[0].Text)
		if text == "" || utf8.RuneCountInString(text) > maxPromptRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid_request: message length must be between 1 and 1200 characters",
			})
			return
		}

		sanitized, err := json.Marshal(env)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal_error: failed to encode request",
			})
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(sanitized))
		r.ContentLength = int64(len(sanitized))
		next(w, r)
	}
}

// clientIP extracts the trailing entry of X-Forwarded-For appended by Cloud Run's GFE,
// falling back to RemoteAddr so callers cannot spoof leading XFF values to evade rate limits.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range slices.Backward(parts) {
			if ip := strings.TrimSpace(part); ip != "" {
				return ip
			}
		}
	}
	if host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
