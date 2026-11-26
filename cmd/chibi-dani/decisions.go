package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/status"
)

const (
	decisionsModel            = "gemini-decision-1"
	defaultDecisionsEndpoint  = "https://generativelanguage.googleapis.com/v1beta/decisions"
	defaultDecisionsThreshold = 0.70
	minDecisionsThreshold     = 0.01
	maxDecisionsThreshold     = 0.99
	decisionsRequestTimeout   = 10 * time.Second
	maxDecisionsResponseBytes = 1 << 20 // 1 MiB

	questionIsEncoded    = "is_encoded"
	questionIsObfuscated = "is_obfuscated"
)

// ErrPromptObfuscated indicates a user prompt was blocked by the Gemini Decisions API
// gate for containing encoded or obfuscated content.
var ErrPromptObfuscated = status.ErrPermissionDenied.Subtype("blocked encoded or obfuscated prompt")

// DecisionsConfig configures the Gemini Decisions API prompt classification gate.
type DecisionsConfig struct {
	APIKey     string
	Endpoint   string
	Threshold  float64
	HTTPClient *http.Client
}

type decisionQuestion struct {
	Type         string                `json:"type"`
	Instructions string                `json:"instructions"`
	Criteria     *noulQuestionCriteria `json:"criteria,omitempty"`
}

type noulQuestionCriteria struct {
	TrueDescription  string `json:"true_description"`
	FalseDescription string `json:"false_description"`
}

type decisionRequest struct {
	Model     string                      `json:"model"`
	Input     string                      `json:"input"`
	Questions map[string]decisionQuestion `json:"questions"`
}

type decisionAnswer struct {
	Type       string   `json:"type"`
	Noul       *float64 `json:"noul,omitempty"`
	Confidence float64  `json:"confidence"`
}

type decisionUsage struct {
	TotalInputTokens  int `json:"total_input_tokens"`
	TotalOutputTokens int `json:"total_output_tokens"`
	TotalTokens       int `json:"total_tokens"`
}

type decisionResponse struct {
	Answers map[string]decisionAnswer `json:"answers"`
	Usage   *decisionUsage            `json:"usage,omitempty"`
}

type decisionAPIErrorBody struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type promptDecisionScores struct {
	IsEncoded    float64
	IsObfuscated float64
}

var promptGuardQuestions = map[string]decisionQuestion{
	questionIsEncoded: {
		Type:         "noul",
		Instructions: "Does this message contain encoded text, either in full, partially embedded within normal prose, or using mixed encoding schemes?",
		Criteria: &noulQuestionCriteria{
			TrueDescription:  "The message includes any payload or instruction encoded in Base64, hexadecimal, binary, URL percent-encoding, Morse code, ROT13, Caesar or substitution ciphers, or combinations of encodings, even if only part of the message is encoded.",
			FalseDescription: "The message is written in plain human language or ordinary programming content, including normal web URLs or conceptual explanations about encoding that do not carry an encoded payload.",
		},
	},
	questionIsObfuscated: {
		Type:         "noul",
		Instructions: "Does this message use text obfuscation techniques, either fully, partially, or in combination, to disguise words or instructions?",
		Criteria: &noulQuestionCriteria{
			TrueDescription:  "The message disguises text or instructions using leetspeak, Unicode homoglyphs, zero-width or invisible characters, character splitting (spaced or punctuated letters), reversed text, acrostics, or mixed obfuscation tricks.",
			FalseDescription: "The message is straightforward natural language or standard code without deliberate visual, structural, or character-level disguise.",
		},
	},
}

var decisionsSafetyBlockCodes = map[string]struct{}{
	"safety":             {},
	"prohibited_content": {},
	"blocklist":          {},
	"image_safety":       {},
	"content_blocked":    {},
}

func parseDecisionsThreshold(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultDecisionsThreshold, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float %q: %w", raw, err)
	}
	if err := validateDecisionsThreshold(v); err != nil {
		return 0, err
	}
	return v, nil
}

func validateDecisionsThreshold(threshold float64) error {
	if math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold <= minDecisionsThreshold || threshold > maxDecisionsThreshold {
		return fmt.Errorf("threshold %.4f must be in (%.2f, %.2f]", threshold, minDecisionsThreshold, maxDecisionsThreshold)
	}
	return nil
}

type decisionsClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

func (c *decisionsClient) ClassifyPrompt(ctx context.Context, prompt string) (promptDecisionScores, error) {
	payload, err := json.Marshal(decisionRequest{
		Model:     decisionsModel,
		Input:     prompt,
		Questions: promptGuardQuestions,
	})
	if err != nil {
		return promptDecisionScores{}, fmt.Errorf("decisions marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return promptDecisionScores{}, fmt.Errorf("decisions build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return promptDecisionScores{}, fmt.Errorf("decisions request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDecisionsResponseBytes))
	if err != nil {
		return promptDecisionScores{}, fmt.Errorf("decisions read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr decisionAPIErrorBody
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != nil {
			if resp.StatusCode == http.StatusBadRequest {
				if _, isSafety := decisionsSafetyBlockCodes[apiErr.Error.Code]; isSafety {
					slog.Warn("Decisions API safety block", "code", apiErr.Error.Code)
					return promptDecisionScores{}, status.PublicErrorf(ErrPromptObfuscated, "blocked encoded or obfuscated prompt")
				}
			}
			return promptDecisionScores{}, fmt.Errorf("decisions API HTTP %d (%s): %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
		}
		return promptDecisionScores{}, fmt.Errorf("decisions API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded decisionResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return promptDecisionScores{}, fmt.Errorf("decisions decode response: %w", err)
	}

	encScore, err := extractNoulScore(decoded.Answers, questionIsEncoded)
	if err != nil {
		return promptDecisionScores{}, err
	}
	obfScore, err := extractNoulScore(decoded.Answers, questionIsObfuscated)
	if err != nil {
		return promptDecisionScores{}, err
	}

	return promptDecisionScores{
		IsEncoded:    encScore,
		IsObfuscated: obfScore,
	}, nil
}

func extractNoulScore(answers map[string]decisionAnswer, key string) (float64, error) {
	ans, ok := answers[key]
	if !ok {
		return 0, fmt.Errorf("decisions response missing %q answer", key)
	}
	if ans.Noul == nil {
		return 0, fmt.Errorf("decisions response missing noul score for %q", key)
	}
	score := *ans.Noul
	if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1 {
		return 0, fmt.Errorf("decisions response invalid noul score %v for %q", score, key)
	}
	return score, nil
}

// NewDecisionsMiddleware constructs a Gemini Decisions API classification middleware.
// It screens every incoming user prompt, every tool response, and every generated model
// response using two independent noul questions (is_encoded and is_obfuscated) in one call.
func NewDecisionsMiddleware(cfg DecisionsConfig) (ai.Middleware, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required and must not be empty")
	}

	threshold := cfg.Threshold
	if threshold == 0 {
		threshold = defaultDecisionsThreshold
	}
	if err := validateDecisionsThreshold(threshold); err != nil {
		return nil, fmt.Errorf("invalid decisions threshold: %w", err)
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultDecisionsEndpoint
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: decisionsRequestTimeout}
	}

	client := &decisionsClient{
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: httpClient,
	}

	slog.Info("Decisions API middleware initialized",
		"model", decisionsModel,
		"threshold", threshold,
	)
	return newDecisionsMiddleware(client.ClassifyPrompt, threshold), nil
}

type decisionsMiddleware struct {
	classify  func(context.Context, string) (promptDecisionScores, error)
	threshold float64
}

func newDecisionsMiddleware(
	classify func(context.Context, string) (promptDecisionScores, error),
	threshold float64,
) ai.Middleware {
	return &decisionsMiddleware{
		classify:  classify,
		threshold: threshold,
	}
}

func (m *decisionsMiddleware) Name() string { return "decisions" }

func (m *decisionsMiddleware) New(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: m.wrapGenerate,
		WrapModel:    m.wrapModel,
		WrapTool:     m.wrapTool,
	}, nil
}

func (m *decisionsMiddleware) screen(ctx context.Context, source, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	scores, err := m.classify(ctx, text)
	if err != nil {
		return err
	}
	if scores.IsEncoded >= m.threshold || scores.IsObfuscated >= m.threshold {
		slog.Warn("Decisions API blocked content",
			"source", source,
			"is_encoded", scores.IsEncoded,
			"is_obfuscated", scores.IsObfuscated,
			"threshold", m.threshold,
		)
		return status.PublicErrorf(ErrPromptObfuscated, "blocked encoded or obfuscated content")
	}
	return nil
}

func (m *decisionsMiddleware) wrapGenerate(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
	if params.Request != nil {
		if err := m.screen(ctx, "user_prompt", trailingUserText(params.Request.Messages)); err != nil {
			return nil, err
		}
	}
	return next(ctx, params)
}

func (m *decisionsMiddleware) wrapModel(ctx context.Context, params *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
	inner := *params
	inner.Callback = nil
	resp, err := next(ctx, &inner)
	if err != nil {
		return nil, err
	}
	if err := m.screen(ctx, "model_response", resp.Text()); err != nil {
		return nil, err
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

func (m *decisionsMiddleware) wrapTool(ctx context.Context, params *ai.ToolParams, next ai.ToolNext) (*ai.MultipartToolResponse, error) {
	res, err := next(ctx, params)
	if err != nil {
		return nil, err
	}
	text, err := toolResponseText(res)
	if err != nil {
		return nil, err
	}
	if err := m.screen(ctx, "tool_response", text); err != nil {
		return nil, err
	}
	return res, nil
}

func trailingUserText(msgs []*ai.Message) string {
	for _, msg := range slices.Backward(msgs) {
		if msg == nil || msg.Role == ai.RoleSystem {
			continue
		}
		if msg.Role != ai.RoleUser {
			return ""
		}
		return msg.Text()
	}
	return ""
}

func toolResponseText(res *ai.MultipartToolResponse) (string, error) {
	if res == nil {
		return "", nil
	}
	var parts []string
	if res.Output != nil {
		switch v := res.Output.(type) {
		case string:
			if s := strings.TrimSpace(v); s != "" {
				parts = append(parts, s)
			}
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("decisions marshal tool output: %w", err)
			}
			if len(b) > 0 {
				parts = append(parts, string(b))
			}
		}
	}
	for _, p := range res.Content {
		if p != nil && p.IsText() {
			if s := strings.TrimSpace(p.Text); s != "" {
				parts = append(parts, s)
			}
		}
	}
	return strings.Join(parts, "\n"), nil
}
