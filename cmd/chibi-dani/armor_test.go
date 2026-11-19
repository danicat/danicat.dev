package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	modelarmorpb "cloud.google.com/go/modelarmor/apiv1/modelarmorpb"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
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

func TestModelArmorMiddlewareScreensInputOnceOnIterationZero(t *testing.T) {
	t.Parallel()

	var screenedInputs []string
	var screenedOutputs []string
	mw := newArmorMiddleware(
		"projects/p/locations/us-central1/templates/t",
		func(_ context.Context, prompt string) error {
			screenedInputs = append(screenedInputs, prompt)
			return nil
		},
		func(_ context.Context, reply string) error {
			screenedOutputs = append(screenedOutputs, reply)
			return nil
		},
		nil,
	)

	hooks, err := mw.New(context.Background())
	if err != nil {
		t.Fatalf("mw.New failed: %v", err)
	}

	req := &ai.ModelRequest{
		Messages: []*ai.Message{
			ai.NewSystemTextMessage("system instructions"),
			ai.NewUserTextMessage("first question"),
			ai.NewModelTextMessage("first answer"),
			ai.NewUserTextMessage("latest user question"),
		},
	}

	// Iteration 0: screens input once; tool-call reply has empty text.
	_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
		Iteration: 0,
		Request:   req,
	}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{Name: "read_corpus"}))}, nil
	})
	if err != nil {
		t.Fatalf("iteration 0 failed: %v", err)
	}

	// Iteration 1 (post-tool step): must NOT re-screen user prompt, MUST screen final reply.
	_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
		Iteration: 1,
		Request:   req,
	}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelTextMessage("grounded answer")}, nil
	})
	if err != nil {
		t.Fatalf("iteration 1 failed: %v", err)
	}

	if len(screenedInputs) != 1 || screenedInputs[0] != "latest user question" {
		t.Fatalf("expected input screened exactly once with latest user prompt, got %#v", screenedInputs)
	}
	if len(screenedOutputs) != 2 || screenedOutputs[1] != "grounded answer" {
		t.Fatalf("expected output hook to see grounded answer, got %#v", screenedOutputs)
	}
}

func TestModelArmorMiddlewareBlocksFlaggedContent(t *testing.T) {
	t.Parallel()

	t.Run("blocks prompt before invoking model", func(t *testing.T) {
		t.Parallel()
		modelCalled := false
		mw := newArmorMiddleware(
			"projects/p/locations/us-central1/templates/t",
			func(_ context.Context, _ string) error {
				return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
			},
			func(_ context.Context, _ string) error { return nil },
			nil,
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

	t.Run("blocks flagged response after inference", func(t *testing.T) {
		t.Parallel()
		mw := newArmorMiddleware(
			"projects/p/locations/us-central1/templates/t",
			func(_ context.Context, _ string) error { return nil },
			func(_ context.Context, _ string) error {
				return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
			},
			nil,
		)
		hooks, err := mw.New(context.Background())
		if err != nil {
			t.Fatalf("mw.New failed: %v", err)
		}

		_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
			Iteration: 0,
			Request:   &ai.ModelRequest{Messages: []*ai.Message{ai.NewUserTextMessage("benign prompt")}},
		}, func(_ context.Context, _ *ai.GenerateParams) (*ai.ModelResponse, error) {
			return &ai.ModelResponse{Message: ai.NewModelTextMessage("unsafe output")}, nil
		})
		if !errors.Is(err, ErrModelArmorBlocked) {
			t.Fatalf("expected ErrModelArmorBlocked, got %v", err)
		}
	})
}

type fakeSanitizeStream struct {
	ctx        context.Context
	resCh      chan *modelarmorpb.SanitizeModelResponseResponse
	blockWord  string
	seenModes  []modelarmorpb.StreamingMode
}

func (f *fakeSanitizeStream) Send(req *modelarmorpb.SanitizeModelResponseRequest) error {
	if err := f.ctx.Err(); err != nil {
		return err
	}
	f.seenModes = append(f.seenModes, req.GetStreamingMode())
	text := req.GetModelResponseData().GetText()
	state := modelarmorpb.FilterMatchState_NO_MATCH_FOUND
	if f.blockWord != "" && strings.Contains(text, f.blockWord) {
		state = modelarmorpb.FilterMatchState_MATCH_FOUND
	}
	select {
	case f.resCh <- &modelarmorpb.SanitizeModelResponseResponse{
		SanitizationResult: &modelarmorpb.SanitizationResult{
			FilterMatchState: state,
			SanitizationMetadata: &modelarmorpb.SanitizationResult_SanitizationMetadata{
				StreamChunkProcessed: &modelarmorpb.DataItem{
					DataItem: &modelarmorpb.DataItem_Text{Text: text},
				},
			},
		},
	}:
		return nil
	case <-f.ctx.Done():
		return f.ctx.Err()
	}
}

func (f *fakeSanitizeStream) Recv() (*modelarmorpb.SanitizeModelResponseResponse, error) {
	select {
	case res, ok := <-f.resCh:
		if !ok {
			return nil, io.EOF
		}
		return res, nil
	case <-f.ctx.Done():
		return nil, f.ctx.Err()
	}
}

func (f *fakeSanitizeStream) CloseSend() error {
	close(f.resCh)
	return nil
}

func TestModelArmorMiddlewareStreamingRealtime(t *testing.T) {
	t.Parallel()

	t.Run("withholds flagged streamed chunk and cancels generation", func(t *testing.T) {
		t.Parallel()

		var fake *fakeSanitizeStream
		mw := newArmorMiddleware(
			"projects/p/locations/us-central1/templates/t",
			func(_ context.Context, _ string) error { return nil },
			func(_ context.Context, _ string) error { return nil },
			func(ctx context.Context) (responseSanitizeStream, error) {
				fake = &fakeSanitizeStream{
					ctx:       ctx,
					resCh:     make(chan *modelarmorpb.SanitizeModelResponseResponse, 8),
					blockWord: "FORBIDDEN",
				}
				return fake, nil
			},
		)
		hooks, err := mw.New(context.Background())
		if err != nil {
			t.Fatalf("mw.New failed: %v", err)
		}

		var delivered []string
		var sawToolChunk bool
		origCB := func(_ context.Context, chunk *ai.ModelResponseChunk) error {
			if chunk.Text() == "" {
				sawToolChunk = true
				return nil
			}
			delivered = append(delivered, chunk.Text())
			return nil
		}

		_, err = hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
			Iteration: 0,
			Request:   &ai.ModelRequest{Messages: []*ai.Message{ai.NewUserTextMessage("hello")}},
			Callback:  origCB,
		}, func(ctx context.Context, p *ai.GenerateParams) (*ai.ModelResponse, error) {
			// Tool chunk should pass through without opening Model Armor stream.
			if err := p.Callback(ctx, &ai.ModelResponseChunk{
				Role:    ai.RoleModel,
				Content: []*ai.Part{ai.NewToolRequestPart(&ai.ToolRequest{Name: "read_corpus"})},
			}); err != nil {
				return nil, err
			}
			if err := p.Callback(ctx, &ai.ModelResponseChunk{
				Role:    ai.RoleModel,
				Content: []*ai.Part{ai.NewTextPart("Safe prefix. ")},
			}); err != nil {
				return nil, err
			}
			if err := p.Callback(ctx, &ai.ModelResponseChunk{
				Role:    ai.RoleModel,
				Content: []*ai.Part{ai.NewTextPart("FORBIDDEN payload")},
			}); err != nil {
				return nil, err
			}
			// Wait for Model Armor recv goroutine to cancel streamCtx.
			<-ctx.Done()
			return nil, ctx.Err()
		})

		if !errors.Is(err, ErrModelArmorBlocked) {
			t.Fatalf("expected ErrModelArmorBlocked, got %v", err)
		}
		if !sawToolChunk {
			t.Fatal("expected tool-call status chunk to pass through to callback")
		}
		if fake == nil || len(fake.seenModes) == 0 || fake.seenModes[0] != modelarmorpb.StreamingMode_STREAMING_MODE_REALTIME {
			t.Fatalf("expected STREAMING_MODE_REALTIME, got %#v", fake)
		}
		joined := strings.Join(delivered, "")
		if strings.Contains(joined, "FORBIDDEN") {
			t.Fatalf("flagged chunk must NEVER be delivered to client callback, got %q", joined)
		}
		if joined != "Safe prefix. " {
			t.Fatalf("expected verified prefix chunk to be delivered, got %q", joined)
		}
	})

	t.Run("streams all verified chunks in order when safe", func(t *testing.T) {
		t.Parallel()

		mw := newArmorMiddleware(
			"projects/p/locations/us-central1/templates/t",
			func(_ context.Context, _ string) error { return nil },
			func(_ context.Context, _ string) error { return nil },
			func(ctx context.Context) (responseSanitizeStream, error) {
				return &fakeSanitizeStream{
					ctx:   ctx,
					resCh: make(chan *modelarmorpb.SanitizeModelResponseResponse, 8),
				}, nil
			},
		)
		hooks, err := mw.New(context.Background())
		if err != nil {
			t.Fatalf("mw.New failed: %v", err)
		}

		var delivered []string
		origCB := func(_ context.Context, chunk *ai.ModelResponseChunk) error {
			delivered = append(delivered, chunk.Text())
			return nil
		}

		resp, err := hooks.WrapGenerate(context.Background(), &ai.GenerateParams{
			Iteration: 0,
			Request:   &ai.ModelRequest{Messages: []*ai.Message{ai.NewUserTextMessage("hello")}},
			Callback:  origCB,
		}, func(ctx context.Context, p *ai.GenerateParams) (*ai.ModelResponse, error) {
			for _, part := range []string{"Part 1. ", "Part 2."} {
				if err := p.Callback(ctx, &ai.ModelResponseChunk{
					Role:    ai.RoleModel,
					Content: []*ai.Part{ai.NewTextPart(part)},
				}); err != nil {
					return nil, err
				}
			}
			return &ai.ModelResponse{Message: ai.NewModelTextMessage("Part 1. Part 2.")}, nil
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp == nil || resp.Text() != "Part 1. Part 2." {
			t.Fatalf("unexpected response: %v", resp)
		}
		if got := strings.Join(delivered, ""); got != "Part 1. Part 2." {
			t.Fatalf("expected full verified stream, got %q", got)
		}
	})
}
