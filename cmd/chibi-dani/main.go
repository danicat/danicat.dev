package main

import (
	"cmp"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	"github.com/firebase/genkit/go/plugins/googlecloud"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	mwexp "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/firebase/genkit/go/plugins/server"
	"google.golang.org/genai"
)

//go:embed system_prompt.md
var baseSystemPrompt string

type readCorpusInput struct {
	Path string `json:"path" jsonschema_description:"Canonical danicat.dev URL or relative path from the Master Content Index (for example https://danicat.dev/posts/<slug>/, https://danicat.dev/about/, https://danicat.dev/events/, or https://danicat.dev/videos/)"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("chibi-dani exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	armorTemplate := strings.TrimSpace(os.Getenv("MODEL_ARMOR_TEMPLATE"))
	if projectID == "" || armorTemplate == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT and MODEL_ARMOR_TEMPLATE are required")
	}
	port := cmp.Or(strings.TrimSpace(os.Getenv("PORT")), "8080")
	corpusBaseURL := cmp.Or(strings.TrimSpace(os.Getenv("CORPUS_BASE_URL")), defaultCorpusBaseURL)
	skillsLLMSURL := cmp.Or(strings.TrimSpace(os.Getenv("SKILLS_LLMS_URL")), defaultSkillsLLMSURL)
	refreshInterval := defaultCorpusRefreshInterval
	if rawRefresh := strings.TrimSpace(os.Getenv("CORPUS_REFRESH_INTERVAL")); rawRefresh != "" {
		parsed, err := time.ParseDuration(rawRefresh)
		if err != nil || parsed <= 0 {
			return fmt.Errorf("invalid CORPUS_REFRESH_INTERVAL %q: must be a positive duration", rawRefresh)
		}
		refreshInterval = parsed
	}

	googlecloud.EnableGoogleCloudTelemetry(&googlecloud.GoogleCloudTelemetryOptions{
		ProjectID: projectID,
	})

	corpusStore, err := NewRemoteCorpusStore(ctx, corpusBaseURL, WithRefreshInterval(refreshInterval), WithSkillsLLMSURL(skillsLLMSURL))
	if err != nil {
		return fmt.Errorf("load corpus: %w", err)
	}

	armorMW, closeArmor, err := NewModelArmorMiddleware(ctx, armorTemplate)
	if err != nil {
		return fmt.Errorf("init model armor: %w", err)
	}
	defer func() {
		if closeErr := closeArmor(); closeErr != nil {
			slog.Warn("Model Armor close warning", "error", closeErr)
		}
	}()

	fsClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("init firestore: %w", err)
	}
	defer func() {
		if closeErr := fsClient.Close(); closeErr != nil {
			slog.Warn("Firestore close warning", "error", closeErr)
		}
	}()

	g := genkit.Init(ctx,
		genkit.WithExperimental(),
		genkit.WithPlugins(&googlegenai.VertexAI{ProjectID: projectID, Location: "global"}),
	)

	readCorpus := genkit.DefineTool(g, "read_corpus",
		"Read the full markdown source of a published danicat.dev page listed in the Master Content Index.",
		func(_ *ai.ToolContext, in readCorpusInput) (string, error) {
			text, ok := corpusStore.Snapshot().Lookup(in.Path)
			if !ok {
				return fmt.Sprintf("page %q not found in published corpus", in.Path), nil
			}
			return text, nil
		},
	)

	sessions := newBoundedSessionStore[any](defaultMaxSessions, defaultSessionTTL)
	chibi := genkitx.DefineAgent(g, "chibi-dani",
		aix.InlinePrompt{
			ai.WithModelName("vertexai/gemini-3.5-flash-lite"),
			ai.WithConfig(&genai.GenerateContentConfig{MaxOutputTokens: 1500}),
			ai.WithSystemFn(func(ctx context.Context, _ any) (string, error) {
				c := corpusStore.Current(ctx)
				return strings.TrimSpace(baseSystemPrompt) + "\n" + c.PromptIndex, nil
			}),
			ai.WithMaxTurns(5),
			ai.WithTools(readCorpus),
			ai.WithUse(
				NewTelemetryMiddleware(fsClient, "chibi_sessions"),
				armorMW,
				&mwexp.Budget{
					Scope: mwexp.BudgetScopeRun,
					Limit: ai.GenerationUsage{InputTokens: 160_000, OutputTokens: 1_500, ThoughtsTokens: 16_000},
				},
				&mwexp.Budget{
					Scope: mwexp.BudgetScopeSession,
					Limit: ai.GenerationUsage{InputTokens: 500_000, OutputTokens: 12_000, ThoughtsTokens: 64_000},
				},
			),
		},
		aix.WithSessionStore[any](sessions),
		aix.WithDescription[any]("Chibi Dani, grounded companion agent for danicat.dev"),
	)

	allowLocalhost := strings.EqualFold(strings.TrimSpace(os.Getenv("CHIBI_ALLOW_LOCALHOST")), "true")
	snap := corpusStore.Snapshot()
	slog.Info("Chibi Dani ready",
		"project", projectID,
		"corpus_base_url", corpusBaseURL,
		"refresh_interval", refreshInterval.String(),
		"posts", snap.PostCount,
		"sections", snap.SectionCount,
		"port", port,
		"allow_localhost", allowLocalhost,
	)

	mux := newHTTPMux(chibi.Name(), sessions, genkit.Handler(chibi), allowLocalhost)
	return server.Start(ctx, ":"+port, mux)
}
