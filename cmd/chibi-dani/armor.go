package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	modelarmor "cloud.google.com/go/modelarmor/apiv1"
	modelarmorpb "cloud.google.com/go/modelarmor/apiv1/modelarmorpb"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
	"google.golang.org/api/option"
)

// ErrModelArmorBlocked indicates a prompt or response tripped Google Cloud Model Armor filters.
var ErrModelArmorBlocked = status.ErrPermissionDenied.Subtype("blocked by Model Armor security policy")

// NewModelArmorMiddleware constructs a Google Cloud Model Armor middleware.
// WrapGenerate screens every incoming user prompt with SanitizeUserPrompt.
// WrapModel withholds streaming until unary SanitizeModelResponse clears each
// model response, then emits the approved chunk to the caller's stream callback.
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

	c := &armorClient{client: client, template: templateName}
	slog.Info("Model Armor middleware initialized", "template", templateName, "location", loc)
	return newArmorMiddleware(c.sanitizeUserPrompt, c.sanitizeModelResponse), client.Close, nil
}

type armorClient struct {
	client   *modelarmor.Client
	template string
}

func (c *armorClient) sanitizeUserPrompt(ctx context.Context, userPrompt string) error {
	userPrompt = strings.TrimSpace(userPrompt)
	if userPrompt == "" {
		return nil
	}
	res, err := c.client.SanitizeUserPrompt(ctx, &modelarmorpb.SanitizeUserPromptRequest{
		Name:                           c.template,
		UserPromptData:                 &modelarmorpb.DataItem{DataItem: &modelarmorpb.DataItem_Text{Text: userPrompt}},
		MultiLanguageDetectionMetadata: &modelarmorpb.MultiLanguageDetectionMetadata{EnableMultiLanguageDetection: true},
	})
	if err != nil {
		return fmt.Errorf("modelarmor SanitizeUserPrompt: %w", err)
	}
	if res.GetSanitizationResult().GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND {
		slog.Warn("Model Armor blocked user input", "template", c.template)
		return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
	}
	return nil
}

func (c *armorClient) sanitizeModelResponse(ctx context.Context, modelResponse string) error {
	modelResponse = strings.TrimSpace(modelResponse)
	if modelResponse == "" {
		return nil
	}
	res, err := c.client.SanitizeModelResponse(ctx, &modelarmorpb.SanitizeModelResponseRequest{
		Name:                           c.template,
		ModelResponseData:              &modelarmorpb.DataItem{DataItem: &modelarmorpb.DataItem_Text{Text: modelResponse}},
		MultiLanguageDetectionMetadata: &modelarmorpb.MultiLanguageDetectionMetadata{EnableMultiLanguageDetection: true},
	})
	if err != nil {
		return fmt.Errorf("modelarmor SanitizeModelResponse: %w", err)
	}
	if res.GetSanitizationResult().GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND {
		slog.Warn("Model Armor blocked model response", "template", c.template)
		return status.PublicErrorf(ErrModelArmorBlocked, "blocked by Model Armor security policy")
	}
	return nil
}

type armorMiddleware struct {
	sanitizeInput  func(context.Context, string) error
	sanitizeOutput func(context.Context, string) error
}

func newArmorMiddleware(
	sanitizeInput func(context.Context, string) error,
	sanitizeOutput func(context.Context, string) error,
) ai.Middleware {
	return &armorMiddleware{
		sanitizeInput:  sanitizeInput,
		sanitizeOutput: sanitizeOutput,
	}
}

func (m *armorMiddleware) Name() string { return "model-armor" }

func (m *armorMiddleware) New(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: m.wrapGenerate,
		WrapModel:    m.wrapModel,
	}, nil
}

func (m *armorMiddleware) wrapGenerate(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
	if params.Request != nil {
		if prompt := strings.TrimSpace(trailingUserText(params.Request.Messages)); prompt != "" {
			if err := m.sanitizeInput(ctx, prompt); err != nil {
				return nil, err
			}
		}
	}
	return next(ctx, params)
}

func (m *armorMiddleware) wrapModel(ctx context.Context, params *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
	inner := *params
	inner.Callback = nil
	resp, err := next(ctx, &inner)
	if err != nil {
		return nil, err
	}
	if text := strings.TrimSpace(resp.Text()); text != "" {
		if err := m.sanitizeOutput(ctx, text); err != nil {
			return nil, err
		}
	}
	if params.Callback != nil && resp != nil && resp.Message != nil && len(resp.Message.Content) > 0 {
		if err := params.Callback(ctx, &ai.ModelResponseChunk{
			Role:    cmp.Or(resp.Message.Role, ai.RoleModel),
			Content: resp.Message.Content,
		}); err != nil {
			return nil, err
		}
	}
	return resp, nil
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
