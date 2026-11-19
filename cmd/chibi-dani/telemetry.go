package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
)

var (
	reEmail        = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	reSecret       = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=\-]{16,}|(?:AIza[0-9A-Za-z_\-]{20,}|ya29\.[0-9A-Za-z._\-]{16,}|gh[pousr]_[0-9A-Za-z]{20,}|sk-[0-9A-Za-z_\-]{16,})`)
	rePhone        = regexp.MustCompile(`(?:\+\d{1,3}[\s.-]?)?\(\d{2,4}\)[\s.-]?\d{3,5}[\s.-]?\d{4}\b|\+\d{1,3}[\s.-]+\d{2,4}[\s.-]+\d{3,5}[\s.-]?\d{4}\b|\b\d{3}[-.\s]\d{3}[-.\s]\d{4}\b`)
	reMarkdownLink = regexp.MustCompile(`\[[^\]\n]+\]\(((?:[^()\s\n]|\([^()\s\n]*\))+)\)`)
)

type turnRecord struct {
	sessionID  string
	turnID     string
	occurredAt time.Time
	question   string
	citedURLs  []string
	latencyMs  int64
	outcome    string
}

// NewTelemetryMiddleware records PII-redacted questions, cited links, latency,
// and outcomes per turn into Cloud Firestore without interfering with SSE streams.
func NewTelemetryMiddleware(fs *firestore.Client, collection string) ai.Middleware {
	return newTelemetryMiddleware(func(ctx context.Context, t turnRecord) error {
		writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()

		_, err := fs.Collection(collection).Doc(t.sessionID).Set(writeCtx, map[string]any{
			"session_id":    t.sessionID,
			"updated_at":    t.occurredAt,
			"turn_count":    firestore.Increment(1),
			"last_question": t.question,
			"last_outcome":  t.outcome,
			"turns": firestore.ArrayUnion(map[string]any{
				"turn_id":       t.turnID,
				"occurred_at":   t.occurredAt,
				"question":      t.question,
				"cited_urls":    t.citedURLs,
				"has_citations": len(t.citedURLs) > 0,
				"latency_ms":    t.latencyMs,
				"outcome":       t.outcome,
			}),
		}, firestore.MergeAll)
		return err
	})
}

func newTelemetryMiddleware(save func(context.Context, turnRecord) error) ai.Middleware {
	return ai.MiddlewareFunc(func(_ context.Context) (*ai.Hooks, error) {
		return &ai.Hooks{
			WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
				if params == nil || params.Iteration > 0 {
					return next(ctx, params)
				}

				start := time.Now()
				var question string
				if params.Request != nil {
					question = RedactPII(latestUserPromptText(params.Request.Messages))
				}

				resp, err := next(ctx, params)

				sess := aix.SessionFromContext[any](ctx)
				if sess == nil || sess.SessionID() == "" {
					slog.Error("telemetry skipped: missing genkit session in context")
					return resp, err
				}

				turnID := fmt.Sprintf("turn-%d", start.UnixNano())
				if tc := aix.TurnContextFromContext(ctx); tc != nil && tc.SnapshotID != "" {
					turnID = tc.SnapshotID
				}

				citedURLs := []string{}
				if err == nil && resp != nil {
					citedURLs = extractMarkdownLinks(resp.Text())
				}

				saveErr := save(ctx, turnRecord{
					sessionID:  sess.SessionID(),
					turnID:     turnID,
					occurredAt: time.Now().UTC(),
					question:   question,
					citedURLs:  citedURLs,
					latencyMs:  time.Since(start).Milliseconds(),
					outcome:    classifyTurnOutcome(err),
				})
				if saveErr != nil {
					slog.Error("firestore telemetry write failed", "session_id", sess.SessionID(), "error", saveErr)
				}
				return resp, err
			},
		}, nil
	})
}

// RedactPII masks email addresses, secret tokens, and phone numbers in logged prompts.
func RedactPII(text string) string {
	out := strings.TrimSpace(text)
	if out == "" {
		return ""
	}
	out = reEmail.ReplaceAllString(out, "[EMAIL_REDACTED]")
	out = reSecret.ReplaceAllString(out, "[SECRET_REDACTED]")
	out = rePhone.ReplaceAllString(out, "[PHONE_REDACTED]")
	return out
}

func extractMarkdownLinks(rawText string) []string {
	matches := reMarkdownLink.FindAllStringSubmatch(rawText, -1)
	seen := make(map[string]struct{}, len(matches))
	urls := make([]string, 0, min(10, len(matches)))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		u := strings.TrimSpace(m[1])
		if u == "" || (!strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://")) {
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		urls = append(urls, u)
		if len(urls) >= 10 {
			break
		}
	}
	return urls
}

func classifyTurnOutcome(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, ErrModelArmorBlocked):
		return "armor_blocked"
	case errors.Is(err, ai.ErrBudgetExceeded):
		return "budget_exceeded"
	default:
		return "error"
	}
}
