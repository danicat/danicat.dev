package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
	"github.com/firebase/genkit/go/genkit"
)

func TestNewModelArmorMiddlewareFailFast(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if _, _, err := NewModelArmorMiddleware(ctx, ""); err == nil {
		t.Fatal("expected error for empty MODEL_ARMOR_TEMPLATE")
	}
	if _, _, err := NewModelArmorMiddleware(ctx, "projects/p/locations/global/templates/t"); err == nil {
		t.Fatal("expected error for global MODEL_ARMOR_TEMPLATE location")
	}
}

func TestModelArmorMiddlewareScreensInputAndModelOutputs(t *testing.T) {
	t.Parallel()

	var screenedInputs []string
	var screenedOutputs []string
	mw := newArmorMiddleware(
		func(_ context.Context, prompt string) error {
			screenedInputs = append(screenedInputs, prompt)
			return nil
		},
		func(_ context.Context, reply string) error {
			screenedOutputs = append(screenedOutputs, reply)
			return nil
		},
	)

	hooks, err := mw.New(context.Background())
	if err != nil {
		t.Fatalf("mw.New failed: %v", err)
	}

	userReq := &ai.ModelRequest{
		Messages: []*ai.Message{
			ai.NewSystemTextMessage("system instructions"),
			ai.NewUserTextMessage("first question"),
			ai.NewModelTextMessage("first answer"),
			ai.NewUserTextMessage("latest user question"),
		},
	}

	// Iteration 0: screens trailing user message once.
	_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
		Iteration: 0,
		Request:   userReq,
	}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelTextMessage("ok")}, nil
	})
	if err != nil {
		t.Fatalf("iteration 0 WrapGenerate failed: %v", err)
	}

	// Iteration 1 (post-tool message list ending in RoleTool): does not re-screen user prompt.
	toolReq := &ai.ModelRequest{
		Messages: append(append([]*ai.Message{}, userReq.Messages...),
			ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{Name: "read_corpus"})),
			ai.NewMessage(ai.RoleTool, nil, ai.NewToolResponsePart(&ai.ToolResponse{Name: "read_corpus", Output: "doc"})),
		),
	}
	_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
		Iteration: 1,
		Request:   toolReq,
	}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelTextMessage("grounded answer")}, nil
	})
	if err != nil {
		t.Fatalf("iteration 1 WrapGenerate failed: %v", err)
	}

	// WrapModel on tool-request-only reply skips empty text; text reply is screened once.
	_, err = hooks.WrapModel(context.Background(), &ai.ModelParams{
		Request: userReq,
	}, func(_ context.Context, _ *ai.ModelParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{Name: "read_corpus"}))}, nil
	})
	if err != nil {
		t.Fatalf("tool-request WrapModel failed: %v", err)
	}

	_, err = hooks.WrapModel(context.Background(), &ai.ModelParams{
		Request: toolReq,
	}, func(_ context.Context, _ *ai.ModelParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelTextMessage("grounded answer")}, nil
	})
	if err != nil {
		t.Fatalf("final WrapModel failed: %v", err)
	}

	if len(screenedInputs) != 1 || screenedInputs[0] != "latest user question" {
		t.Fatalf("expected input screened once with latest user prompt, got %#v", screenedInputs)
	}
	if len(screenedOutputs) != 1 || screenedOutputs[0] != "grounded answer" {
		t.Fatalf("expected output screened once with grounded answer, got %#v", screenedOutputs)
	}
}

func TestModelArmorMiddlewareBlocksFlaggedContent(t *testing.T) {
	t.Parallel()

	t.Run("blocks prompt before invoking model", func(t *testing.T) {
		t.Parallel()
		modelCalled := false
		mw := newArmorMiddleware(
			func(_ context.Context, _ string) error {
				return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
			},
			func(_ context.Context, _ string) error { return nil },
		)
		hooks, err := mw.New(context.Background())
		if err != nil {
			t.Fatalf("mw.New failed: %v", err)
		}

		_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
			Iteration: 0,
			Request:   &ai.ModelRequest{Messages: []*ai.Message{ai.NewUserTextMessage("jailbreak")}},
		}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
			modelCalled = true
			return &ai.ModelResponse{Message: ai.NewModelTextMessage("should not run")}, nil
		})
		if !errors.Is(err, ErrModelArmorBlocked) {
			t.Fatalf("expected ErrModelArmorBlocked, got %v", err)
		}
		if modelCalled {
			t.Fatal("model must not be called when Model Armor blocks the prompt")
		}
	})

	t.Run("blocks flagged response in WrapModel and leaks zero chunks", func(t *testing.T) {
		t.Parallel()
		mw := newArmorMiddleware(
			func(_ context.Context, _ string) error { return nil },
			func(_ context.Context, _ string) error {
				return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
			},
		)
		hooks, err := mw.New(context.Background())
		if err != nil {
			t.Fatalf("mw.New failed: %v", err)
		}

		var leaked []string
		_, err = hooks.WrapModel(context.Background(), &ai.ModelParams{
			Request: &ai.ModelRequest{Messages: []*ai.Message{ai.NewUserTextMessage("benign prompt")}},
			Callback: func(_ context.Context, chunk *ai.ModelResponseChunk) error {
				leaked = append(leaked, chunk.Text())
				return nil
			},
		}, func(_ context.Context, p *ai.ModelParams) (*ai.ModelResponse, error) {
			if p.Callback != nil {
				t.Fatal("expected WrapModel to clear inner Callback before invoking model")
			}
			return &ai.ModelResponse{Message: ai.NewModelTextMessage("unsafe output")}, nil
		})
		if !errors.Is(err, ErrModelArmorBlocked) {
			t.Fatalf("expected ErrModelArmorBlocked, got %v", err)
		}
		if len(leaked) != 0 {
			t.Fatalf("must not leak blocked chunks, got %#v", leaked)
		}
	})
}

func TestModelArmorMiddlewareEndToEndStreaming(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	g := genkit.Init(ctx, genkit.WithExperimental())

	var sawCallbackMu sync.Mutex
	var sawNonNilCallback bool
	genkit.DefineModelAction(g, "test/armor-stream-mock",
		&ai.ModelOptions{
			Supports: &ai.ModelSupports{Multiturn: true},
		},
		func(_ context.Context, req *ai.ModelRequest, _ map[string]any, cb ai.ModelStreamCallback) (*ai.ModelResponse, error) {
			if cb != nil {
				sawCallbackMu.Lock()
				sawNonNilCallback = true
				sawCallbackMu.Unlock()
			}
			reply := "Safe grounded answer."
			if strings.Contains(latestUserPromptText(req.Messages), "trigger-unsafe") {
				reply = "FORBIDDEN payload"
			}
			return &ai.ModelResponse{
				FinishReason: ai.FinishReasonStop,
				Message:      ai.NewModelTextMessage(reply),
			}, nil
		},
	)

	mw := newArmorMiddleware(
		func(_ context.Context, _ string) error { return nil },
		func(_ context.Context, text string) error {
			if strings.Contains(text, "FORBIDDEN") {
				return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
			}
			return nil
		},
	)

	t.Run("withholds flagged response through real Genkit streaming", func(t *testing.T) {
		var delivered []string
		_, err := genkit.Generate(ctx, g,
			ai.WithModelName("test/armor-stream-mock"),
			ai.WithPrompt("trigger-unsafe"),
			ai.WithUse(mw),
			ai.WithStreaming(func(_ context.Context, chunk *ai.ModelResponseChunk) error {
				delivered = append(delivered, chunk.Text())
				return nil
			}),
		)
		if !errors.Is(err, ErrModelArmorBlocked) {
			t.Fatalf("expected ErrModelArmorBlocked, got %v", err)
		}
		if len(delivered) != 0 {
			t.Fatalf("must deliver zero chunks when Model Armor blocks, got %#v", delivered)
		}
	})

	t.Run("emits approved chunk through real Genkit streaming", func(t *testing.T) {
		var delivered []string
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName("test/armor-stream-mock"),
			ai.WithPrompt("normal question"),
			ai.WithUse(mw),
			ai.WithStreaming(func(_ context.Context, chunk *ai.ModelResponseChunk) error {
				delivered = append(delivered, chunk.Text())
				return nil
			}),
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.Text() != "Safe grounded answer." {
			t.Fatalf("unexpected response text: %q", resp.Text())
		}
		if got := strings.Join(delivered, ""); got != "Safe grounded answer." {
			t.Fatalf("expected approved streamed chunk, got %q", got)
		}
	})

	sawCallbackMu.Lock()
	defer sawCallbackMu.Unlock()
	if sawNonNilCallback {
		t.Fatal("model action must receive nil callback while WrapModel withholds streaming")
	}
}
