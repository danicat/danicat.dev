package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core/status"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	mwexp "github.com/firebase/genkit/go/plugins/middleware/exp"
)

type sseResultEnvelope struct {
	Result *struct {
		SessionID    string `json:"sessionId"`
		FinishReason string `json:"finishReason"`
		Error        *struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
		Message *struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	} `json:"result"`
	Error *struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"error"`
}

func postSSE(t *testing.T, handler http.Handler, sessionID, text string) sseResultEnvelope {
	t.Helper()

	payload := map[string]any{
		"data": map[string]any{
			"message": map[string]any{
				"role":    "user",
				"content": []map[string]string{{"text": text}},
			},
		},
	}
	if sessionID != "" {
		payload["init"] = map[string]string{"sessionId": sessionID}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/agents/chibi-dani?stream=true", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Origin", "https://danicat.dev")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 SSE stream, got %d: %s", rec.Code, rec.Body.String())
	}

	var lastResult sseResultEnvelope
	var foundResult bool
	for line := range strings.SplitSeq(rec.Body.String(), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
		if dataStr == "" || dataStr == "[DONE]" {
			continue
		}
		var env sseResultEnvelope
		if err := json.Unmarshal([]byte(dataStr), &env); err != nil {
			t.Fatalf("decode SSE event %q: %v", dataStr, err)
		}
		if env.Error != nil {
			t.Fatalf("unexpected top-level SSE error event: %+v", env.Error)
		}
		if env.Result != nil {
			lastResult = env
			foundResult = true
		}
	}
	if !foundResult {
		t.Fatalf("no final result SSE event in stream:\n%s", rec.Body.String())
	}
	return lastResult
}

func TestPublicBudgetMiddlewareNilFailsFast(t *testing.T) {
	t.Parallel()

	mw := newPublicBudgetMiddleware(nil)
	if _, err := mw.New(context.Background()); !errors.Is(err, status.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for nil budget, got %v", err)
	}
}

func TestBudgetAndMaxTurnsEndToEndProductionSSE(t *testing.T) {
	// Force production mode so genkit.Handler applies wire redaction to non-public errors.
	t.Setenv("GENKIT_ENV", "")

	t.Run("session budget exhaustion returns public error over SSE across turns", func(t *testing.T) {
		ctx := context.Background()
		g := genkit.Init(ctx, genkit.WithExperimental())

		var modelCalls atomic.Int32
		genkit.DefineModelAction(g, "test/session-budget-mock",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Multiturn: true}},
			func(_ context.Context, _ *ai.ModelRequest, _ map[string]any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				modelCalls.Add(1)
				return &ai.ModelResponse{
					FinishReason: ai.FinishReasonStop,
					Message:      ai.NewModelTextMessage("Grounded answer."),
					Usage:        &ai.GenerationUsage{InputTokens: 60, OutputTokens: 20},
				}, nil
			},
		)

		var mu sync.Mutex
		var outcomes []string
		telemetryMW := newTelemetryMiddleware(func(_ context.Context, r turnRecord) error {
			mu.Lock()
			outcomes = append(outcomes, r.outcome)
			mu.Unlock()
			return nil
		})

		sessions := newBoundedSessionStore[any](defaultMaxSessions, defaultSessionTTL)
		agent := genkitx.DefineAgent(g, "chibi-dani",
			aix.InlinePrompt{
				ai.WithModelName("test/session-budget-mock"),
				ai.WithUse(
					telemetryMW,
					newPublicMaxTurnsMiddleware(),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeSession,
						Limit: ai.GenerationUsage{InputTokens: 100},
					}),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeRun,
						Limit: ai.GenerationUsage{InputTokens: 160_000},
					}),
				),
			},
			aix.WithSessionStore[any](sessions),
		)

		mux := newHTTPMuxWithLimiters(
			agent.Name(),
			sessions,
			genkit.Handler(agent),
			buildAllowedOrigins(false),
			newIPRateLimiter(64, time.Millisecond, 100),
			newGlobalGuard(time.Millisecond, 100, 1000, time.Hour, time.Now),
		)

		// Turn 1: 60 / 100 input tokens.
		res1 := postSSE(t, mux, "", "First question")
		if res1.Result.Error != nil || res1.Result.FinishReason != "stop" {
			t.Fatalf("turn 1 unexpected result: %+v", res1.Result)
		}
		sessionID := res1.Result.SessionID
		if sessionID == "" {
			t.Fatal("expected non-empty sessionId from turn 1")
		}

		// Turn 2: finishes at 120 / 100 input tokens (budget checks happen before each model call).
		res2 := postSSE(t, mux, sessionID, "Second question")
		if res2.Result.Error != nil || res2.Result.FinishReason != "stop" {
			t.Fatalf("turn 2 unexpected result: %+v", res2.Result)
		}

		// Turn 3: session is now over budget; must stop before calling the model with public message.
		beforeCalls := modelCalls.Load()
		res3 := postSSE(t, mux, sessionID, "why daniela hasn't written a book yet?")
		if modelCalls.Load() != beforeCalls {
			t.Fatalf("model must not be called once session budget is exhausted (calls %d -> %d)", beforeCalls, modelCalls.Load())
		}
		if res3.Result.FinishReason != "aborted" {
			t.Fatalf("turn 3 finishReason = %q, want %q", res3.Result.FinishReason, "aborted")
		}
		if res3.Result.Error == nil {
			t.Fatal("turn 3 expected non-nil error")
		}
		if res3.Result.Error.Status != string(status.Aborted) {
			t.Fatalf("turn 3 error status = %q, want %q", res3.Result.Error.Status, status.Aborted)
		}
		if res3.Result.Error.Message != sessionBudgetPublicMessage {
			t.Fatalf("turn 3 error message = %q, want %q", res3.Result.Error.Message, sessionBudgetPublicMessage)
		}
		if strings.Contains(res3.Result.Error.Message, "120") || strings.Contains(res3.Result.Error.Message, "100") {
			t.Fatalf("turn 3 public error leaked internal token counts: %q", res3.Result.Error.Message)
		}

		// Turn 4 (follow-up like "hello?"): stays blocked with the same public session budget message.
		res4 := postSSE(t, mux, sessionID, "hello?")
		if modelCalls.Load() != beforeCalls {
			t.Fatalf("turn 4 must not call model after session exhaustion")
		}
		if res4.Result.Error == nil || res4.Result.Error.Message != sessionBudgetPublicMessage {
			t.Fatalf("turn 4 error = %+v, want %q", res4.Result.Error, sessionBudgetPublicMessage)
		}

		mu.Lock()
		gotOutcomes := slices.Clone(outcomes)
		mu.Unlock()
		wantOutcomes := []string{"ok", "ok", "budget_exceeded", "budget_exceeded"}
		if !slices.Equal(gotOutcomes, wantOutcomes) {
			t.Fatalf("telemetry outcomes = %v, want %v", gotOutcomes, wantOutcomes)
		}
	})

	t.Run("run budget stops single turn while leaving session usable", func(t *testing.T) {
		ctx := context.Background()
		g := genkit.Init(ctx, genkit.WithExperimental())

		lookup := genkit.DefineTool(g, "read_corpus", "Read an article",
			func(_ *ai.ToolContext, in map[string]string) (string, error) {
				return "content for " + in["path"], nil
			},
		)

		genkit.DefineModelAction(g, "test/run-budget-mock",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true}},
			func(_ context.Context, req *ai.ModelRequest, _ map[string]any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				if strings.Contains(latestUserPromptText(req.Messages), "heavy") {
					return &ai.ModelResponse{
						FinishReason: ai.FinishReasonStop,
						Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{
							Name:  "read_corpus",
							Input: map[string]any{"path": "big.md"},
						})),
						Usage: &ai.GenerationUsage{InputTokens: 60},
					}, nil
				}
				return &ai.ModelResponse{
					FinishReason: ai.FinishReasonStop,
					Message:      ai.NewModelTextMessage("Quick answer."),
					Usage:        &ai.GenerationUsage{InputTokens: 10},
				}, nil
			},
		)

		sessions := newBoundedSessionStore[any](defaultMaxSessions, defaultSessionTTL)
		agent := genkitx.DefineAgent(g, "chibi-dani",
			aix.InlinePrompt{
				ai.WithModelName("test/run-budget-mock"),
				ai.WithTools(lookup),
				ai.WithMaxTurns(5),
				ai.WithUse(
					newPublicMaxTurnsMiddleware(),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeSession,
						Limit: ai.GenerationUsage{InputTokens: 500},
					}),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeRun,
						Limit: ai.GenerationUsage{InputTokens: 50},
					}),
				),
			},
			aix.WithSessionStore[any](sessions),
		)

		mux := newHTTPMuxWithLimiters(
			agent.Name(),
			sessions,
			genkit.Handler(agent),
			buildAllowedOrigins(false),
			newIPRateLimiter(64, time.Millisecond, 100),
			newGlobalGuard(time.Millisecond, 100, 1000, time.Hour, time.Now),
		)

		// Turn 1: succeeds and establishes session.
		res1 := postSSE(t, mux, "", "light question")
		if res1.Result.Error != nil || res1.Result.FinishReason != "stop" {
			t.Fatalf("turn 1 unexpected result: %+v", res1.Result)
		}
		sessionID := res1.Result.SessionID

		// Turn 2: heavy tool loop exceeds 50-token run cap (60 >= 50) while session is only 70 / 500.
		res2 := postSSE(t, mux, sessionID, "heavy question")
		if res2.Result.FinishReason != "aborted" {
			t.Fatalf("turn 2 finishReason = %q, want aborted", res2.Result.FinishReason)
		}
		if res2.Result.Error == nil || res2.Result.Error.Message != runBudgetPublicMessage {
			t.Fatalf("turn 2 error = %+v, want %q", res2.Result.Error, runBudgetPublicMessage)
		}

		// Turn 3: narrower question in the same session still works.
		res3 := postSSE(t, mux, sessionID, "another light question")
		if res3.Result.Error != nil || res3.Result.FinishReason != "stop" {
			t.Fatalf("turn 3 expected success after run budget stop, got %+v", res3.Result)
		}
	})

	t.Run("session budget wins over run budget when both exceed on same turn", func(t *testing.T) {
		ctx := context.Background()
		g := genkit.Init(ctx, genkit.WithExperimental())

		lookup := genkit.DefineTool(g, "read_corpus", "Read an article",
			func(_ *ai.ToolContext, in map[string]string) (string, error) {
				return "content", nil
			},
		)

		genkit.DefineModelAction(g, "test/both-budgets-mock",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true}},
			func(_ context.Context, _ *ai.ModelRequest, _ map[string]any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				return &ai.ModelResponse{
					FinishReason: ai.FinishReasonStop,
					Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{
						Name:  "read_corpus",
						Input: map[string]any{"path": "big.md"},
					})),
					Usage: &ai.GenerationUsage{InputTokens: 200},
				}, nil
			},
		)

		sessions := newBoundedSessionStore[any](defaultMaxSessions, defaultSessionTTL)
		agent := genkitx.DefineAgent(g, "chibi-dani",
			aix.InlinePrompt{
				ai.WithModelName("test/both-budgets-mock"),
				ai.WithTools(lookup),
				ai.WithMaxTurns(5),
				ai.WithUse(
					newPublicMaxTurnsMiddleware(),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeSession,
						Limit: ai.GenerationUsage{InputTokens: 100},
					}),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeRun,
						Limit: ai.GenerationUsage{InputTokens: 50},
					}),
				),
			},
			aix.WithSessionStore[any](sessions),
		)

		mux := newHTTPMuxWithLimiters(
			agent.Name(),
			sessions,
			genkit.Handler(agent),
			buildAllowedOrigins(false),
			newIPRateLimiter(64, time.Millisecond, 100),
			newGlobalGuard(time.Millisecond, 100, 1000, time.Hour, time.Now),
		)

		res := postSSE(t, mux, "", "overflow both")
		if res.Result.FinishReason != "aborted" {
			t.Fatalf("finishReason = %q, want aborted", res.Result.FinishReason)
		}
		if res.Result.Error == nil || res.Result.Error.Message != sessionBudgetPublicMessage {
			t.Fatalf("error = %+v, want %q", res.Result.Error, sessionBudgetPublicMessage)
		}
	})

	t.Run("max turns exceeded returns distinct public error over SSE", func(t *testing.T) {
		ctx := context.Background()
		g := genkit.Init(ctx, genkit.WithExperimental())

		lookup := genkit.DefineTool(g, "read_corpus", "Read an article",
			func(_ *ai.ToolContext, in map[string]string) (string, error) {
				return "content", nil
			},
		)

		genkit.DefineModelAction(g, "test/max-turns-mock",
			&ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true}},
			func(_ context.Context, _ *ai.ModelRequest, _ map[string]any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				return &ai.ModelResponse{
					FinishReason: ai.FinishReasonStop,
					Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{
						Name:  "read_corpus",
						Input: map[string]any{"path": "loop.md"},
					})),
					Usage: &ai.GenerationUsage{InputTokens: 10},
				}, nil
			},
		)

		var mu sync.Mutex
		var outcomes []string
		telemetryMW := newTelemetryMiddleware(func(_ context.Context, r turnRecord) error {
			mu.Lock()
			outcomes = append(outcomes, r.outcome)
			mu.Unlock()
			return nil
		})

		sessions := newBoundedSessionStore[any](defaultMaxSessions, defaultSessionTTL)
		agent := genkitx.DefineAgent(g, "chibi-dani",
			aix.InlinePrompt{
				ai.WithModelName("test/max-turns-mock"),
				ai.WithTools(lookup),
				ai.WithMaxTurns(2),
				ai.WithUse(
					telemetryMW,
					newPublicMaxTurnsMiddleware(),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeSession,
						Limit: ai.GenerationUsage{InputTokens: 500_000},
					}),
					newPublicBudgetMiddleware(&mwexp.Budget{
						Scope: mwexp.BudgetScopeRun,
						Limit: ai.GenerationUsage{InputTokens: 160_000},
					}),
				),
			},
			aix.WithSessionStore[any](sessions),
		)

		mux := newHTTPMuxWithLimiters(
			agent.Name(),
			sessions,
			genkit.Handler(agent),
			buildAllowedOrigins(false),
			newIPRateLimiter(64, time.Millisecond, 100),
			newGlobalGuard(time.Millisecond, 100, 1000, time.Hour, time.Now),
		)

		res := postSSE(t, mux, "", "loop forever")
		if res.Result.FinishReason != "aborted" {
			t.Fatalf("finishReason = %q, want aborted", res.Result.FinishReason)
		}
		if res.Result.Error == nil || res.Result.Error.Message != maxTurnsPublicMessage {
			t.Fatalf("error = %+v, want %q", res.Result.Error, maxTurnsPublicMessage)
		}

		mu.Lock()
		gotOutcomes := slices.Clone(outcomes)
		mu.Unlock()
		if !slices.Equal(gotOutcomes, []string{"max_turns_exceeded"}) {
			t.Fatalf("telemetry outcomes = %v, want [max_turns_exceeded]", gotOutcomes)
		}
	})
}
