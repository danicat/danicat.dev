package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"

	modelarmor "cloud.google.com/go/modelarmor/apiv1"
	modelarmorpb "cloud.google.com/go/modelarmor/apiv1/modelarmorpb"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
	"google.golang.org/api/option"
)

// ErrModelArmorBlocked indicates a prompt or response tripped Google Cloud Model Armor filters.
var ErrModelArmorBlocked = status.ErrPermissionDenied.Subtype("blocked by Model Armor security policy")

type responseSanitizeStream interface {
	Send(*modelarmorpb.SanitizeModelResponseRequest) error
	Recv() (*modelarmorpb.SanitizeModelResponseResponse, error)
	CloseSend() error
}

// NewModelArmorMiddleware constructs a Google Cloud Model Armor middleware.
// SanitizeUserPrompt screens the latest user message once per turn (Iteration == 0).
// Non-streaming responses use unary SanitizeModelResponse; streaming SSE responses
// use StreamSanitizeModelResponse in STREAMING_MODE_REALTIME so unchecked chunks
// never reach the client.
func NewModelArmorMiddleware(ctx context.Context, templateName string) (ai.Middleware, func() error, error) {
	templateName = strings.TrimSpace(templateName)
	if templateName == "" {
		return nil, nil, fmt.Errorf("MODEL_ARMOR_TEMPLATE is required and must not be empty")
	}

	loc := extractTemplateSegment(templateName, "locations")
	if loc == "" || loc == "global" {
		return nil, nil, fmt.Errorf("MODEL_ARMOR_TEMPLATE %q must specify a regional location (for example us-central1)", templateName)
	}

	client, err := modelarmor.NewClient(ctx,
		option.WithEndpoint(fmt.Sprintf("modelarmor.%s.rep.googleapis.com:443", loc)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("modelarmor.NewClient(%q): %w", templateName, err)
	}

	sanitizeInput := func(ctx context.Context, userPrompt string) error {
		if userPrompt == "" {
			return nil
		}
		res, err := client.SanitizeUserPrompt(ctx, &modelarmorpb.SanitizeUserPromptRequest{
			Name:                           templateName,
			UserPromptData:                 &modelarmorpb.DataItem{DataItem: &modelarmorpb.DataItem_Text{Text: userPrompt}},
			MultiLanguageDetectionMetadata: &modelarmorpb.MultiLanguageDetectionMetadata{EnableMultiLanguageDetection: true},
		})
		if err != nil {
			return fmt.Errorf("modelarmor SanitizeUserPrompt: %w", err)
		}
		if res.GetSanitizationResult().GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND {
			slog.Warn("Model Armor blocked user input", "template", templateName)
			return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
		}
		return nil
	}

	sanitizeOutput := func(ctx context.Context, modelResponse string) error {
		modelResponse = strings.TrimSpace(modelResponse)
		if modelResponse == "" {
			return nil
		}
		res, err := client.SanitizeModelResponse(ctx, &modelarmorpb.SanitizeModelResponseRequest{
			Name:                           templateName,
			ModelResponseData:              &modelarmorpb.DataItem{DataItem: &modelarmorpb.DataItem_Text{Text: modelResponse}},
			MultiLanguageDetectionMetadata: &modelarmorpb.MultiLanguageDetectionMetadata{EnableMultiLanguageDetection: true},
		})
		if err != nil {
			return fmt.Errorf("modelarmor SanitizeModelResponse: %w", err)
		}
		if res.GetSanitizationResult().GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND {
			slog.Warn("Model Armor blocked model response", "template", templateName)
			return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
		}
		return nil
	}

	openStream := func(ctx context.Context) (responseSanitizeStream, error) {
		return client.StreamSanitizeModelResponse(ctx)
	}

	slog.Info("Model Armor middleware initialized", "template", templateName, "location", loc)
	return newArmorMiddleware(templateName, sanitizeInput, sanitizeOutput, openStream), client.Close, nil
}

func newArmorMiddleware(
	templateName string,
	sanitizeInput func(context.Context, string) error,
	sanitizeOutput func(context.Context, string) error,
	openStream func(context.Context) (responseSanitizeStream, error),
) ai.Middleware {
	return ai.MiddlewareFunc(func(_ context.Context) (*ai.Hooks, error) {
		return &ai.Hooks{
			WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
				if params == nil {
					return next(ctx, params)
				}

				if params.Iteration == 0 && params.Request != nil {
					if err := sanitizeInput(ctx, latestUserPromptText(params.Request.Messages)); err != nil {
						return nil, err
					}
				}

				if params.Callback == nil {
					resp, err := next(ctx, params)
					if err != nil {
						return nil, err
					}
					if resp != nil {
						if err := sanitizeOutput(ctx, resp.Text()); err != nil {
							return nil, err
						}
					}
					return resp, nil
				}

				origCB := params.Callback
				streamCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				var (
					stream   responseSanitizeStream
					recvDone chan struct{}
					cbMu     sync.Mutex
					errMu    sync.Mutex
					recvErr  error
				)
				setRecvErr := func(e error) {
					errMu.Lock()
					if recvErr == nil {
						recvErr = e
					}
					errMu.Unlock()
				}
				getRecvErr := func() error {
					errMu.Lock()
					defer errMu.Unlock()
					return recvErr
				}

				mode := modelarmorpb.StreamingMode_STREAMING_MODE_REALTIME
				params.Callback = func(_ context.Context, chunk *ai.ModelResponseChunk) error {
					if chunk == nil {
						return nil
					}
					if err := getRecvErr(); err != nil {
						return err
					}
					text := chunk.Text()
					if text == "" {
						// Forward non-text status chunks (such as tool-call notifications) directly.
						cbMu.Lock()
						err := origCB(streamCtx, chunk)
						cbMu.Unlock()
						return err
					}

					if stream == nil {
						s, err := openStream(streamCtx)
						if err != nil {
							return fmt.Errorf("modelarmor StreamSanitizeModelResponse open: %w", err)
						}
						stream = s
						recvDone = make(chan struct{})
						chunkIndex := chunk.Index
						go func() {
							defer close(recvDone)
							for {
								res, err := s.Recv()
								if err == io.EOF {
									return
								}
								if err != nil {
									if streamCtx.Err() == nil {
										setRecvErr(fmt.Errorf("modelarmor StreamSanitizeModelResponse recv: %w", err))
										cancel()
									}
									return
								}
								sr := res.GetSanitizationResult()
								if sr.GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND {
									slog.Warn("Model Armor blocked streamed model response", "template", templateName)
									setRecvErr(status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy"))
									cancel()
									return
								}
								if verified := sr.GetSanitizationMetadata().GetStreamChunkProcessed().GetText(); verified != "" {
									cbMu.Lock()
									cbErr := origCB(streamCtx, &ai.ModelResponseChunk{
										Role:    ai.RoleModel,
										Index:   chunkIndex,
										Content: []*ai.Part{ai.NewTextPart(verified)},
									})
									cbMu.Unlock()
									if cbErr != nil {
										setRecvErr(cbErr)
										cancel()
										return
									}
								}
							}
						}()
					}

					if err := stream.Send(&modelarmorpb.SanitizeModelResponseRequest{
						Name: templateName,
						ModelResponseData: &modelarmorpb.DataItem{
							DataItem: &modelarmorpb.DataItem_Text{Text: text},
						},
						MultiLanguageDetectionMetadata: &modelarmorpb.MultiLanguageDetectionMetadata{
							EnableMultiLanguageDetection: true,
						},
						StreamingMode: &mode,
					}); err != nil {
						if rErr := getRecvErr(); rErr != nil {
							return rErr
						}
						return fmt.Errorf("modelarmor StreamSanitizeModelResponse send: %w", err)
					}
					return nil
				}

				resp, err := next(streamCtx, params)
				if stream != nil {
					if err != nil {
						cancel()
					}
					_ = stream.CloseSend()
					<-recvDone
					if rErr := getRecvErr(); rErr != nil {
						return nil, rErr
					}
				}
				if err != nil {
					return nil, err
				}
				if stream == nil && resp != nil && strings.TrimSpace(resp.Text()) != "" {
					if err := sanitizeOutput(ctx, resp.Text()); err != nil {
						return nil, err
					}
				}
				return resp, nil
			},
		}, nil
	})
}

func latestUserPromptText(msgs []*ai.Message) string {
	for _, msg := range slices.Backward(msgs) {
		if msg != nil && msg.Role == ai.RoleUser {
			if text := strings.TrimSpace(msg.Text()); text != "" {
				return text
			}
		}
	}
	return ""
}

func extractTemplateSegment(templateName, key string) string {
	parts := strings.Split(templateName, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == key {
			return strings.TrimSpace(parts[i+1])
		}
	}
	return ""
}
