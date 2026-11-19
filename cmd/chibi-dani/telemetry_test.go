package main

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
)

func TestRedactPII(t *testing.T) {
	t.Parallel()
	in := "Email daniela@example.com or call +55 11 98765-4321 using Bearer sk-proj-1234567890abcdef on 2026-10-06 (Genkit v1.14.0)"
	got := RedactPII(in)

	for _, secret := range []string{"daniela@example.com", "98765-4321", "sk-proj-1234567890abcdef"} {
		if strings.Contains(got, secret) {
			t.Fatalf("expected %q to be redacted, got %q", secret, got)
		}
	}
	for _, preserved := range []string{"[EMAIL_REDACTED]", "[PHONE_REDACTED]", "[SECRET_REDACTED]", "2026-10-06", "v1.14.0"} {
		if !strings.Contains(got, preserved) {
			t.Fatalf("expected %q to remain in %q", preserved, got)
		}
	}
}

func TestTelemetryMiddlewareMultiTurn(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	g := genkit.Init(ctx, genkit.WithExperimental())

	corpusReadTool := genkit.DefineTool(g, "read_corpus", "Read a published article",
		func(_ *ai.ToolContext, input map[string]string) (string, error) {
			return "Article summary for " + input["path"], nil
		},
	)

	var callSeq int
	genkit.DefineModelAction(g, "test/chibi-mock",
		&ai.ModelOptions{
			Supports: &ai.ModelSupports{
				Tools:     true,
				Multiturn: true,
			},
		},
		func(_ context.Context, _ *ai.ModelRequest, _ map[string]any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
			callSeq++
			if callSeq == 1 {
				return &ai.ModelResponse{
					FinishReason: ai.FinishReasonStop,
					Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{
						Name:  "read_corpus",
						Input: map[string]any{"path": "posts/20260216-genkit-go/index.md"},
					})),
				}, nil
			}
			return &ai.ModelResponse{
				FinishReason: ai.FinishReasonStop,
				Message:      ai.NewModelTextMessage("Read [Genkit Go](/posts/20260216-genkit-go/) and https://cloud.google.com/run."),
			}, nil
		},
	)

	var mu sync.Mutex
	var recorded []turnRecord
	mw := newTelemetryMiddleware(func(_ context.Context, r turnRecord) error {
		mu.Lock()
		defer mu.Unlock()
		recorded = append(recorded, r)
		return nil
	})

	agent := genkitx.DefineAgent(g, "test-chibi",
		aix.InlinePrompt{
			ai.WithModelName("test/chibi-mock"),
			ai.WithTools(corpusReadTool),
			ai.WithUse(mw),
		},
		aix.WithSessionStore[any](localstore.NewInMemorySessionStore[any]()),
	)

	out1, err := agent.RunText(ctx, "Tell me about Genkit Go, my email is user@example.com")
	if err != nil || out1.FinishReason == aix.AgentFinishReasonFailed {
		t.Fatalf("turn 1 failed: err=%v outErr=%v", err, out1)
	}
	out2, err := agent.RunText(ctx, "How do I deploy it to Cloud Run?", aix.WithSessionID[any](out1.SessionID))
	if err != nil || out2.FinishReason == aix.AgentFinishReasonFailed {
		t.Fatalf("turn 2 failed: err=%v outErr=%v", err, out2)
	}
	if out2.SessionID != out1.SessionID {
		t.Fatalf("expected shared session ID %q, got %q", out1.SessionID, out2.SessionID)
	}

	mu.Lock()
	turns := slices.Clone(recorded)
	mu.Unlock()
	if len(turns) != 2 {
		t.Fatalf("expected 2 recorded turns (1 per outer turn even with tool loop), got %d", len(turns))
	}

	first := turns[0]
	if first.sessionID != out1.SessionID || first.outcome != "ok" {
		t.Fatalf("unexpected turn 1 record: %+v", first)
	}
	if !strings.Contains(first.question, "[EMAIL_REDACTED]") || strings.Contains(first.question, "user@example.com") {
		t.Fatalf("expected redacted question on turn 1, got %q", first.question)
	}
	if len(first.citedURLs) != 1 || first.citedURLs[0] != "/posts/20260216-genkit-go/" {
		t.Fatalf("expected extracted markdown link, got %v", first.citedURLs)
	}

	if turns[1].sessionID != out1.SessionID {
		t.Fatalf("expected turn 2 on same session %q, got %q", out1.SessionID, turns[1].sessionID)
	}
}
