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

	"cloud.google.com/go/firestore"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	"github.com/firebase/genkit/go/plugins/googlecloud"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	mwexp "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/firebase/genkit/go/plugins/server"
)

//go:embed system_prompt.md
var baseSystemPrompt string

type readCorpusInput struct {
	Path string `json:"path" jsonschema_description:"Exact file path from the Master Content Index"`
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
	corpusDir := cmp.Or(strings.TrimSpace(os.Getenv("CORPUS_DIR")), ".")

	googlecloud.EnableGoogleCloudTelemetry(&googlecloud.GoogleCloudTelemetryOptions{
		ProjectID:                    projectID,
		DisableLoggingInputAndOutput: true,
	})

	corpus, err := LoadCorpus(corpusDir)
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
		"Read the full markdown or JSON source of a published danicat.dev file listed in the Master Content Index.",
		func(_ *ai.ToolContext, in readCorpusInput) (string, error) {
			text, ok := corpus.Files[strings.TrimPrefix(strings.TrimSpace(in.Path), "/")]
			if !ok {
				return fmt.Sprintf("file %q not found in published corpus", in.Path), nil
			}
			return text, nil
		},
	)

	sessions := localstore.NewInMemorySessionStore[any]()
	chibi := genkitx.DefineAgent(g, "chibi-dani",
		aix.InlinePrompt{
			ai.WithModelName("vertexai/gemini-3.5-flash-lite"),
			ai.WithSystem(strings.TrimSpace(baseSystemPrompt) + "\n" + corpus.PromptIndex),
			ai.WithMaxTurns(5),
			ai.WithTools(readCorpus),
			ai.WithUse(
				NewTelemetryMiddleware(fsClient, "chibi_sessions"),
				armorMW,
				&mwexp.Budget{
					Scope: mwexp.BudgetScopeRun,
					Limit: ai.GenerationUsage{InputTokens: 160_000, OutputTokens: 4_000, ThoughtsTokens: 16_000},
				},
				&mwexp.Budget{
					Scope: mwexp.BudgetScopeSession,
					Limit: ai.GenerationUsage{InputTokens: 500_000, OutputTokens: 20_000, ThoughtsTokens: 64_000},
				},
			),
		},
		aix.WithSessionStore[any](sessions),
		aix.WithDescription[any]("Chibi Dani, grounded companion agent for danicat.dev"),
	)

	slog.Info("Chibi Dani ready",
		"project", projectID,
		"posts", corpus.PostCount,
		"codelabs", corpus.CodelabCount,
		"port", port,
	)

	mux := newHTTPMux(chibi.Name(), sessions, genkit.Handler(chibi))
	return server.Start(ctx, ":"+port, mux)
}
