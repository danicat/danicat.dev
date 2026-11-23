package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type corpusFixtures struct {
	llms       string
	enFull     string
	ptbrFull   string
	jaFull     string
	etagSuffix string
}

func makeFixtures(suffix, extraPostSlug, extraPostTitle string) corpusFixtures {
	extraLLMS := ""
	extraEN := ""
	extraPTBR := ""
	extraJA := ""
	if extraPostSlug != "" {
		extraLLMS = fmt.Sprintf("- [%s](https://danicat.dev/posts/%s/index.md): Fresh post summary. (2026-11-23 · Category: Agent Development)\n", extraPostTitle, extraPostSlug)
		extraEN = fmt.Sprintf("\n# %s\n\n- **Markdown URL**: https://danicat.dev/posts/%s/index.md\n- **HTML URL**: https://danicat.dev/posts/%s/\n- **Date**: 2026-11-23\n\nNew English content.\n\n---\n", extraPostTitle, extraPostSlug, extraPostSlug)
		extraPTBR = fmt.Sprintf("\n# %s (PT-BR)\n\n- **Markdown URL**: https://danicat.dev/pt-br/posts/%s/index.md\n- **HTML URL**: https://danicat.dev/pt-br/posts/%s/\n- **Date**: 2026-11-23\n\nNovo conteudo em portugues.\n\n---\n", extraPostTitle, extraPostSlug, extraPostSlug)
		extraJA = fmt.Sprintf("\n# %s (JA)\n\n- **Markdown URL**: https://danicat.dev/ja/posts/%s/index.md\n- **HTML URL**: https://danicat.dev/ja/posts/%s/\n- **Date**: 2026-11-23\n\n新しい日本語の記事。\n\n---\n", extraPostTitle, extraPostSlug, extraPostSlug)
	}

	llms := "# danicat.dev — Daniela Petruzalek\n\n" +
		"## Site Sections & Overview\n\n" +
		"- [Homepage — danicat.dev](https://danicat.dev/index.md): Overview.\n" +
		"- [About](https://danicat.dev/about/index.md): Bio and open-source tools.\n\n" +
		"## Articles & Deep Dives\n" +
		extraLLMS +
		"- [Pragmatic Guide to Agent Skills](https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/index.md): Current skills guide. (2026-08-29 · Category: Agentic Coding)\n" +
		"- [Mastering Agent Skills in Gemini CLI](https://danicat.dev/posts/agent-skills-gemini-cli/index.md) [DEPRECATED]: Old CLI skills guide. (2026-01-29 · Category: Agentic Coding · Superseded by: [Pragmatic Guide to Agent Skills](https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/index.md))\n"

	enFull := "# danicat.dev — Daniela Petruzalek (Full Corpus)\n\n---\n\n" +
		"# About\n\n- **Markdown URL**: https://danicat.dev/about/index.md\n- **HTML URL**: https://danicat.dev/about/\n\nGo and AI engineer.\n\n---\n\n" +
		"# Interactive Codelabs & Workshops\n\n- **HTML URL**: https://danicat.dev/codelabs/\n- **Sample Codelab**: Build an agent.\n\n---\n\n" +
		"# Talks & Conference Events\n\n- **HTML URL**: https://danicat.dev/events/\n- **GopherCon Talk**: MCP in Go.\n\n---\n\n" +
		"# Videos & Screencasts\n\n- **HTML URL**: https://danicat.dev/videos/\n- **Demo Video**: Chibi Dani walkthrough.\n\n---" +
		extraEN +
		"\n# Pragmatic Guide to Agent Skills\n\n- **Markdown URL**: https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/index.md\n- **HTML URL**: https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/\n- **Date**: 2026-08-29\n\nFull guide text.\n\n---\n\n" +
		"# Mastering Agent Skills in Gemini CLI [DEPRECATED]\n\n- **Markdown URL**: https://danicat.dev/posts/agent-skills-gemini-cli/index.md\n- **HTML URL**: https://danicat.dev/posts/agent-skills-gemini-cli/\n- **Date**: 2026-01-29\n\nLegacy CLI text.\n\n---\n"

	ptbrFull := "# danicat.dev — PT-BR\n\n---\n\n" +
		"# Sobre\n\n- **Markdown URL**: https://danicat.dev/pt-br/about/index.md\n- **HTML URL**: https://danicat.dev/pt-br/about/\n\nEngenheira de Go e IA.\n\n---\n\n" +
		"# Codelabs\n\n- **HTML URL**: https://danicat.dev/pt-br/codelabs/\n- **Codelab PT-BR**.\n\n---\n\n" +
		"# Eventos\n\n- **HTML URL**: https://danicat.dev/pt-br/events/\n- **Palestra PT-BR**.\n\n---\n\n" +
		"# Videos\n\n- **HTML URL**: https://danicat.dev/pt-br/videos/\n- **Video PT-BR**.\n\n---" +
		extraPTBR +
		"\n# Guia Pratico de Agent Skills\n\n- **Markdown URL**: https://danicat.dev/pt-br/posts/the-pragmatic-guide-to-agent-skills/index.md\n- **HTML URL**: https://danicat.dev/pt-br/posts/the-pragmatic-guide-to-agent-skills/\n\nTexto completo em portugues.\n\n---\n"

	jaFull := "# danicat.dev — JA\n\n---\n\n" +
		"# プロフィール\n\n- **Markdown URL**: https://danicat.dev/ja/about/index.md\n- **HTML URL**: https://danicat.dev/ja/about/\n\nGoとAIのエンジニア。\n\n---\n\n" +
		"# コードラボ\n\n- **HTML URL**: https://danicat.dev/ja/codelabs/\n- **Codelab JA**.\n\n---\n\n" +
		"# イベント\n\n- **HTML URL**: https://danicat.dev/ja/events/\n- **Talk JA**: Session in Japanese.\n\n---\n\n" +
		"# ビデオ\n\n- **HTML URL**: https://danicat.dev/ja/videos/\n- **Video JA**.\n\n---" +
		extraJA +
		"\n# Agent Skills実践ガイド\n\n- **Markdown URL**: https://danicat.dev/ja/posts/the-pragmatic-guide-to-agent-skills/index.md\n- **HTML URL**: https://danicat.dev/ja/posts/the-pragmatic-guide-to-agent-skills/\n\n日本語の全文。\n\n---\n"

	return corpusFixtures{
		llms:       llms,
		enFull:     enFull,
		ptbrFull:   ptbrFull,
		jaFull:     jaFull,
		etagSuffix: suffix,
	}
}

func TestRemoteCorpusStorePreloadAndLookup(t *testing.T) {
	t.Parallel()

	fix := makeFixtures("v1", "", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/llms.txt":
			_, _ = w.Write([]byte(fix.llms))
		case "/llms-full.txt":
			_, _ = w.Write([]byte(fix.enFull))
		case "/pt-br/llms-full.txt":
			_, _ = w.Write([]byte(fix.ptbrFull))
		case "/ja/llms-full.txt":
			_, _ = w.Write([]byte(fix.jaFull))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	store, err := NewRemoteCorpusStore(context.Background(), srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatalf("NewRemoteCorpusStore failed: %v", err)
	}

	snap := store.Snapshot()
	if snap.PostCount != 2 {
		t.Fatalf("expected 2 English posts, got %d", snap.PostCount)
	}
	if snap.SectionCount != 16 {
		t.Fatalf("expected 16 total sections (6 en + 5 pt-br + 5 ja), got %d", snap.SectionCount)
	}

	if strings.Contains(snap.PromptIndex, "/index.md)") {
		t.Fatalf("expected PromptIndex to rewrite /index.md) links to HTML URLs, got:\n%s", snap.PromptIndex)
	}
	if !strings.Contains(snap.PromptIndex, "[Pragmatic Guide to Agent Skills](https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/)") {
		t.Fatalf("expected canonical HTML post link in PromptIndex, got:\n%s", snap.PromptIndex)
	}
	if !strings.Contains(snap.PromptIndex, "[DEPRECATED]") {
		t.Fatalf("expected [DEPRECATED] marker preserved in PromptIndex")
	}
	if !strings.Contains(snap.PromptIndex, "Superseded by: [Pragmatic Guide to Agent Skills](https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/)") {
		t.Fatalf("expected Superseded by HTML link preserved in PromptIndex")
	}

	lookupCases := []struct {
		query string
		want  string
	}{
		{"https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/", "Full guide text."},
		{"https://danicat.dev/posts/the-pragmatic-guide-to-agent-skills/index.md", "Full guide text."},
		{"/posts/the-pragmatic-guide-to-agent-skills/", "Full guide text."},
		{"posts/the-pragmatic-guide-to-agent-skills", "Full guide text."},
		{"https://danicat.dev/pt-br/posts/the-pragmatic-guide-to-agent-skills/", "Texto completo em portugues."},
		{"pt-br/posts/the-pragmatic-guide-to-agent-skills", "Texto completo em portugues."},
		{"https://danicat.dev/ja/posts/the-pragmatic-guide-to-agent-skills/", "日本語の全文。"},
		{"https://danicat.dev/about/", "Go and AI engineer."},
		{"about/index.md", "Go and AI engineer."},
		{"https://danicat.dev/codelabs/", "Sample Codelab"},
		{"https://danicat.dev/events/", "GopherCon Talk"},
		{"https://danicat.dev/videos/", "Chibi Dani walkthrough."},
		{"https://danicat.dev/pt-br/about/", "Engenheira de Go e IA."},
		{"https://danicat.dev/ja/events/", "Session in Japanese."},
	}
	for _, tc := range lookupCases {
		got, ok := snap.Lookup(tc.query)
		if !ok {
			t.Fatalf("Lookup(%q) returned ok=false", tc.query)
		}
		if !strings.Contains(got, tc.want) {
			t.Fatalf("Lookup(%q) = %q, want substring %q", tc.query, got, tc.want)
		}
		if strings.HasSuffix(got, "---") {
			t.Fatalf("Lookup(%q) left trailing separator ---: %q", tc.query, got)
		}
	}

	if _, ok := snap.Lookup("https://danicat.dev/posts/non-existent/"); ok {
		t.Fatalf("expected Lookup for missing post to return false")
	}
}

func TestRemoteCorpusStoreDailyRefreshAndETag(t *testing.T) {
	t.Parallel()

	v1 := makeFixtures("v1", "", "")
	v2 := makeFixtures("v2", "new-go-post", "Brand New Go Article")

	var (
		mu          sync.Mutex
		mode        = "v1"
		currentNow  = time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
		sawIfMatch  []string
		totalReqs   atomic.Int64
	)

	clock := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return currentNow
	}
	advance := func(d time.Duration) {
		mu.Lock()
		currentNow = currentNow.Add(d)
		mu.Unlock()
	}
	setMode := func(m string) {
		mu.Lock()
		mode = m
		mu.Unlock()
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalReqs.Add(1)
		mu.Lock()
		curMode := mode
		if inm := r.Header.Get("If-None-Match"); inm != "" {
			sawIfMatch = append(sawIfMatch, r.URL.Path+"="+inm)
		}
		mu.Unlock()

		if curMode == "fail" {
			http.Error(w, "upstream unavailable", http.StatusInternalServerError)
			return
		}

		fix := v1
		if curMode == "v2" {
			fix = v2
		}
		etag := fmt.Sprintf(`"%s-%s"`, r.URL.Path, fix.etagSuffix)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		switch r.URL.Path {
		case "/llms.txt":
			_, _ = w.Write([]byte(fix.llms))
		case "/llms-full.txt":
			_, _ = w.Write([]byte(fix.enFull))
		case "/pt-br/llms-full.txt":
			_, _ = w.Write([]byte(fix.ptbrFull))
		case "/ja/llms-full.txt":
			_, _ = w.Write([]byte(fix.jaFull))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	store, err := NewRemoteCorpusStore(ctx, srv.URL,
		WithHTTPClient(srv.Client()),
		WithRefreshInterval(24*time.Hour),
		WithRetryBackoff(5*time.Minute),
		WithNow(clock),
	)
	if err != nil {
		t.Fatalf("NewRemoteCorpusStore failed: %v", err)
	}
	if got := totalReqs.Load(); got != 4 {
		t.Fatalf("expected 4 preload requests, got %d", got)
	}
	snap1 := store.Snapshot()

	// Before 24h elapses, Current makes 0 HTTP requests and keeps snap1.
	advance(23 * time.Hour)
	if got := store.Current(ctx); got != snap1 {
		t.Fatalf("expected same snapshot pointer before refresh interval")
	}
	if got := totalReqs.Load(); got != 4 {
		t.Fatalf("expected 0 extra requests before 24h, got total %d", got)
	}

	// Past 24h with matching ETag -> 304 Not Modified preserves exact snapshot pointer.
	advance(2 * time.Hour)
	if got := store.Current(ctx); got != snap1 {
		t.Fatalf("expected 304 Not Modified to preserve existing snapshot pointer")
	}
	if got := totalReqs.Load(); got != 8 {
		t.Fatalf("expected 4 conditional requests after 24h, got total %d", got)
	}
	mu.Lock()
	if len(sawIfMatch) != 4 {
		mu.Unlock()
		t.Fatalf("expected 4 If-None-Match headers, got %v", sawIfMatch)
	}
	mu.Unlock()

	// Next day fails with HTTP 500 -> keep snap1 and apply 5m retry backoff.
	setMode("fail")
	advance(24 * time.Hour)
	if got := store.Current(ctx); got != snap1 {
		t.Fatalf("expected failed refresh to keep previous snapshot")
	}
	if got := totalReqs.Load(); got != 12 {
		t.Fatalf("expected 4 requests during failed refresh, got total %d", got)
	}

	// Within 5m retry backoff, no new requests are sent.
	advance(4 * time.Minute)
	if got := store.Current(ctx); got != snap1 {
		t.Fatalf("expected cached snapshot during retry backoff")
	}
	if got := totalReqs.Load(); got != 12 {
		t.Fatalf("expected 0 requests during retry backoff, got total %d", got)
	}

	// Past 5m backoff with v2 live -> hot-swaps PromptIndex and Files.
	setMode("v2")
	advance(2 * time.Minute)
	snap2 := store.Current(ctx)
	if snap2 == snap1 {
		t.Fatalf("expected new snapshot pointer after 200 OK corpus update")
	}
	if snap2.PostCount != 3 {
		t.Fatalf("expected 3 English posts after v2 refresh, got %d", snap2.PostCount)
	}
	if !strings.Contains(snap2.PromptIndex, "Brand New Go Article") {
		t.Fatalf("expected PromptIndex to include v2 article")
	}
	body, ok := snap2.Lookup("https://danicat.dev/posts/new-go-post/")
	if !ok || !strings.Contains(body, "New English content.") {
		t.Fatalf("expected Lookup for new v2 post to succeed, got ok=%v body=%q", ok, body)
	}
	jaBody, ok := snap2.Lookup("https://danicat.dev/ja/posts/new-go-post/")
	if !ok || !strings.Contains(jaBody, "新しい日本語の記事。") {
		t.Fatalf("expected Lookup for new v2 Japanese post to succeed, got ok=%v body=%q", ok, jaBody)
	}
}

func TestRemoteCorpusStoreStartupFailsFast(t *testing.T) {
	t.Parallel()

	fix := makeFixtures("v1", "", "")

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "http 500 on japanese corpus",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/llms.txt":
					_, _ = w.Write([]byte(fix.llms))
				case "/llms-full.txt":
					_, _ = w.Write([]byte(fix.enFull))
				case "/pt-br/llms-full.txt":
					_, _ = w.Write([]byte(fix.ptbrFull))
				case "/ja/llms-full.txt":
					http.Error(w, "boom", http.StatusInternalServerError)
				}
			},
		},
		{
			name: "zero sections parsed in llms-full.txt",
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/llms.txt":
					_, _ = w.Write([]byte(fix.llms))
				case "/llms-full.txt":
					_, _ = w.Write([]byte("plain text with no sections"))
				case "/pt-br/llms-full.txt":
					_, _ = w.Write([]byte(fix.ptbrFull))
				case "/ja/llms-full.txt":
					_, _ = w.Write([]byte(fix.jaFull))
				}
			},
		},
		{
			name: "unexpected 304 without cached body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotModified)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			if _, err := NewRemoteCorpusStore(context.Background(), srv.URL, WithHTTPClient(srv.Client())); err == nil {
				t.Fatalf("expected NewRemoteCorpusStore to fail fast for %s", tc.name)
			}
		})
	}
}

func TestRemoteCorpusStoreRedirectPolicy(t *testing.T) {
	t.Parallel()

	fix := makeFixtures("v1", "redirected-post", "Redirected Article")

	t.Run("rejects cross-origin redirect", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/llms.txt":
				http.Redirect(w, r, "https://evil.example/llms.txt", http.StatusFound)
			case "/llms-full.txt":
				_, _ = w.Write([]byte(fix.enFull))
			case "/pt-br/llms-full.txt":
				_, _ = w.Write([]byte(fix.ptbrFull))
			case "/ja/llms-full.txt":
				_, _ = w.Write([]byte(fix.jaFull))
			}
		}))
		defer srv.Close()

		_, err := NewRemoteCorpusStore(context.Background(), srv.URL, WithHTTPClient(srv.Client()))
		if err == nil {
			t.Fatalf("expected cross-origin redirect to fail")
		}
		if !strings.Contains(err.Error(), "refused redirect") {
			t.Fatalf("expected error mentioning \"refused redirect\", got: %v", err)
		}
	})

	t.Run("allows same-origin redirect", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/llms.txt":
				http.Redirect(w, r, "/v2/llms.txt", http.StatusFound)
			case "/v2/llms.txt":
				_, _ = w.Write([]byte(fix.llms))
			case "/llms-full.txt":
				_, _ = w.Write([]byte(fix.enFull))
			case "/pt-br/llms-full.txt":
				_, _ = w.Write([]byte(fix.ptbrFull))
			case "/ja/llms-full.txt":
				_, _ = w.Write([]byte(fix.jaFull))
			default:
				http.NotFound(w, r)
			}
		}))
		defer srv.Close()

		store, err := NewRemoteCorpusStore(context.Background(), srv.URL, WithHTTPClient(srv.Client()))
		if err != nil {
			t.Fatalf("expected same-origin redirect to succeed, got: %v", err)
		}
		if !strings.Contains(store.Snapshot().PromptIndex, "Redirected Article") {
			t.Fatalf("expected redirected llms.txt content to be loaded")
		}
	})
}

func TestSystemPromptHygiene(t *testing.T) {
	t.Parallel()

	if strings.Contains(baseSystemPrompt, "—") {
		t.Fatalf("system_prompt.md must not contain em dashes")
	}
	for _, stale := range []string{"data/talks.json", "data/videos.json", "Files:"} {
		if strings.Contains(baseSystemPrompt, stale) {
			t.Fatalf("system_prompt.md contains stale local-file reference %q", stale)
		}
	}
}
