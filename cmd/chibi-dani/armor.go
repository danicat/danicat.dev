package main

import (
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
// SanitizeUserPrompt screens the latest user message before inference, and
// SanitizeModelResponse screens the model reply before returning it.
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

	slog.Info("Model Armor middleware initialized", "template", templateName, "location", loc)
	return newArmorMiddleware(sanitizeInput, sanitizeOutput), client.Close, nil
}

func newArmorMiddleware(
	sanitizeInput func(context.Context, string) error,
	sanitizeOutput func(context.Context, string) error,
) ai.Middleware {
	return ai.MiddlewareFunc(func(_ context.Context) (*ai.Hooks, error) {
		return &ai.Hooks{
			WrapModel: func(ctx context.Context, params *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
				if err := sanitizeInput(ctx, latestUserPromptText(params.Request.Messages)); err != nil {
					return nil, err
				}

				resp, err := next(ctx, params)
				if err != nil {
					return nil, err
				}

				if err := sanitizeOutput(ctx, resp.Text()); err != nil {
					return nil, err
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
