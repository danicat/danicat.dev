package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	maxBodyBytes   = 8192
	maxPromptRunes = 1200
	rateInterval   = 10 * time.Second // 6 requests/min steady state
	rateBurst      = 4
	rateEntryTTL   = 15 * time.Minute
)

var allowedOrigins = map[string]struct{}{
	"https://danicat.dev":     {},
	"https://www.danicat.dev": {},
	"http://localhost:1313":   {},
	"http://127.0.0.1:1313":   {},
}

type visitorBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorBucket
}

func newIPRateLimiter() *ipRateLimiter {
	return &ipRateLimiter{visitors: make(map[string]*visitorBucket)}
}

func (l *ipRateLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if len(l.visitors) > 2048 {
		for k, v := range l.visitors {
			if now.Sub(v.lastSeen) > rateEntryTTL {
				delete(l.visitors, k)
			}
		}
	}

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitorBucket{limiter: rate.NewLimiter(rate.Every(rateInterval), rateBurst)}
		l.visitors[ip] = v
	}
	v.lastSeen = now
	return v.limiter.Allow()
}

func newHTTPMux(agentName string, sessions aix.SnapshotMetadataReader[any], agentHandler http.HandlerFunc) *http.ServeMux {
	cop := http.NewCrossOriginProtection()
	for origin := range allowedOrigins {
		if err := cop.AddTrustedOrigin(origin); err != nil {
			panic(fmt.Sprintf("invalid trusted origin %q: %v", origin, err))
		}
	}
	limiter := newIPRateLimiter()
	agentPath := "/agents/" + agentName

	mux := http.NewServeMux()
	mux.Handle("POST "+agentPath, cop.Handler(withCORSHeaders(
		withRateLimit(limiter, withStrictAgentInput(sessions, agentHandler)),
	)))
	mux.HandleFunc("OPTIONS "+agentPath, withCORSHeaders(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	return mux
}

func withCORSHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
			}
		}
		next(w, r)
	}
}

func withRateLimit(limiter *ipRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(clientIP(r)) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "rate_limited: too many requests, please wait a moment",
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
