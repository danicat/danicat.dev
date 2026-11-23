package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultCorpusBaseURL         = "https://danicat.dev"
	defaultCorpusRefreshInterval = 24 * time.Hour
	defaultCorpusRetryBackoff    = 5 * time.Minute
	defaultCorpusFetchTimeout    = 20 * time.Second
	maxCorpusFileBytes           = 16 << 20
)

var (
	corpusPaths = []string{
		"/llms.txt",
		"/llms-full.txt",
		"/pt-br/llms-full.txt",
		"/ja/llms-full.txt",
	}
	llmsFullSectionRe = regexp.MustCompile(`(?m)^# [^\n]+\n\n- \*\*(?:Markdown|HTML) URL\*\*: (https?://[^\s]+)`)
)

// Corpus holds an immutable snapshot of published danicat.dev content and its prompt index.
type Corpus struct {
	Files        map[string]string
	PromptIndex  string
	PostCount    int
	SectionCount int
	LoadedAt     time.Time
}

// Lookup resolves a canonical URL, markdown URL, or relative path against the corpus snapshot.
func (c *Corpus) Lookup(raw string) (string, bool) {
	if c == nil {
		return "", false
	}
	key := normalizeCorpusKey(raw)
	if key == "" {
		return "", false
	}
	text, ok := c.Files[key]
	return text, ok
}

type cachedResource struct {
	etag         string
	lastModified string
	body         string
}

type fetchResult struct {
	path        string
	notModified bool
	entry       cachedResource
	err         error
}

// CorpusOption configures a remote CorpusStore.
type CorpusOption func(*CorpusStore)

// WithHTTPClient overrides the HTTP client used to fetch corpus endpoints.
func WithHTTPClient(client *http.Client) CorpusOption {
	return func(s *CorpusStore) {
		if client != nil {
			s.client = client
		}
	}
}

// WithRefreshInterval overrides how often the corpus checks for updates.
func WithRefreshInterval(interval time.Duration) CorpusOption {
	return func(s *CorpusStore) {
		if interval > 0 {
			s.refreshInterval = interval
		}
	}
}

// WithRetryBackoff overrides how long to wait before retrying a failed background refresh.
func WithRetryBackoff(backoff time.Duration) CorpusOption {
	return func(s *CorpusStore) {
		if backoff > 0 {
			s.retryBackoff = backoff
		}
	}
}

// WithNow overrides the clock function used for staleness checks and timestamps.
func WithNow(now func() time.Time) CorpusOption {
	return func(s *CorpusStore) {
		if now != nil {
			s.now = now
		}
	}
}

// CorpusStore maintains a concurrency-safe, ETag-cached in-memory Corpus loaded from danicat.dev.
type CorpusStore struct {
	baseURL           string
	client            *http.Client
	refreshInterval   time.Duration
	retryBackoff      time.Duration
	now               func() time.Time
	current           atomic.Pointer[Corpus]
	nextCheckUnixNano atomic.Int64
	mu                sync.Mutex
	cached            map[string]cachedResource
}

// NewRemoteCorpusStore preloads llms.txt and localized llms-full.txt bundles from baseURL.
func NewRemoteCorpusStore(ctx context.Context, baseURL string, opts ...CorpusOption) (*CorpusStore, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return nil, fmt.Errorf("corpus base URL must not be empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("invalid corpus base URL %q", baseURL)
	}

	s := &CorpusStore{
		baseURL:         trimmed,
		client:          &http.Client{Timeout: defaultCorpusFetchTimeout},
		refreshInterval: defaultCorpusRefreshInterval,
		retryBackoff:    defaultCorpusRetryBackoff,
		now:             time.Now,
		cached:          make(map[string]cachedResource, len(corpusPaths)),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	clientCopy := *s.client
	clientCopy.CheckRedirect = sameOriginCheckRedirect(parsed.Scheme, parsed.Host)
	s.client = &clientCopy

	initCtx, cancel := context.WithTimeout(ctx, defaultCorpusFetchTimeout)
	defer cancel()
	if _, err := s.refresh(initCtx); err != nil {
		return nil, fmt.Errorf("preload remote corpus: %w", err)
	}
	s.nextCheckUnixNano.Store(s.now().Add(s.refreshInterval).UnixNano())
	return s, nil
}

func sameOriginCheckRedirect(allowedScheme, allowedHost string) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("stopped after 5 redirects")
		}
		if !strings.EqualFold(req.URL.Scheme, allowedScheme) || !strings.EqualFold(req.URL.Host, allowedHost) {
			return fmt.Errorf("refused redirect to disallowed target %s", req.URL.Redacted())
		}
		return nil
	}
}

// Snapshot returns the currently loaded Corpus snapshot without checking staleness.
func (s *CorpusStore) Snapshot() *Corpus {
	return s.current.Load()
}

// Current checks whether the refresh interval has elapsed, refreshes via ETag if needed,
// and returns the active Corpus snapshot.
func (s *CorpusStore) Current(ctx context.Context) *Corpus {
	s.RefreshIfStale(ctx)
	return s.current.Load()
}

// RefreshIfStale conditionally refreshes the corpus when the refresh interval has elapsed.
// When upstream returns 304 Not Modified, the existing snapshot pointer is preserved.
// If a refresh fails, the last known good snapshot stays active until retryBackoff elapses.
func (s *CorpusStore) RefreshIfStale(ctx context.Context) {
	if s.now().UnixNano() < s.nextCheckUnixNano.Load() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.now().UnixNano() < s.nextCheckUnixNano.Load() {
		return
	}

	refreshCtx, cancel := context.WithTimeout(ctx, defaultCorpusFetchTimeout)
	defer cancel()

	changed, err := s.refresh(refreshCtx)
	if err != nil {
		slog.Error("corpus_refresh_failed", "error", err, "retry_in", s.retryBackoff)
		s.nextCheckUnixNano.Store(s.now().Add(s.retryBackoff).UnixNano())
		return
	}

	s.nextCheckUnixNano.Store(s.now().Add(s.refreshInterval).UnixNano())
	snap := s.current.Load()
	if !changed {
		slog.Info("corpus_refresh_unchanged", "posts", snap.PostCount, "sections", snap.SectionCount)
		return
	}
	slog.Info("corpus_refreshed", "posts", snap.PostCount, "sections", snap.SectionCount)
}

// refresh fetches all corpus paths concurrently. Caller must hold s.mu (or be in construction).
func (s *CorpusStore) refresh(ctx context.Context) (bool, error) {
	results := make([]fetchResult, len(corpusPaths))
	var wg sync.WaitGroup
	for i, path := range corpusPaths {
		prev, hasPrev := s.cached[path]
		wg.Add(1)
		go func(idx int, p string, cached cachedResource, ok bool) {
			defer wg.Done()
			results[idx] = s.fetchPath(ctx, p, cached, ok)
		}(i, path, prev, hasPrev)
	}
	wg.Wait()

	allNotModified := true
	updatedCache := make(map[string]cachedResource, len(corpusPaths))
	for _, res := range results {
		if res.err != nil {
			return false, res.err
		}
		if !res.notModified {
			allNotModified = false
		}
		updatedCache[res.path] = res.entry
	}

	if allNotModified && s.current.Load() != nil {
		return false, nil
	}

	next, err := s.buildCorpus(updatedCache)
	if err != nil {
		return false, err
	}
	s.cached = updatedCache
	s.current.Store(next)
	return true, nil
}

func (s *CorpusStore) fetchPath(ctx context.Context, path string, prev cachedResource, hasPrev bool) fetchResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return fetchResult{path: path, err: fmt.Errorf("build request for %s: %w", path, err)}
	}
	if hasPrev && prev.body != "" {
		if prev.etag != "" {
			req.Header.Set("If-None-Match", prev.etag)
		} else if prev.lastModified != "" {
			req.Header.Set("If-Modified-Since", prev.lastModified)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fetchResult{path: path, err: fmt.Errorf("fetch %s: %w", path, err)}
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		if !hasPrev || prev.body == "" {
			return fetchResult{path: path, err: fmt.Errorf("unexpected 304 Not Modified without cached body for %s", path)}
		}
		return fetchResult{path: path, notModified: true, entry: prev}
	case http.StatusOK:
		data, err := io.ReadAll(io.LimitReader(resp.Body, maxCorpusFileBytes+1))
		if err != nil {
			return fetchResult{path: path, err: fmt.Errorf("read %s: %w", path, err)}
		}
		if len(data) > maxCorpusFileBytes {
			return fetchResult{path: path, err: fmt.Errorf("response for %s exceeds %d byte limit", path, maxCorpusFileBytes)}
		}
		return fetchResult{
			path: path,
			entry: cachedResource{
				etag:         strings.TrimSpace(resp.Header.Get("ETag")),
				lastModified: strings.TrimSpace(resp.Header.Get("Last-Modified")),
				body:         string(data),
			},
		}
	default:
		return fetchResult{path: path, err: fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, path)}
	}
}

func (s *CorpusStore) buildCorpus(resources map[string]cachedResource) (*Corpus, error) {
	llmsTxt := resources["/llms.txt"].body
	if strings.TrimSpace(llmsTxt) == "" {
		return nil, fmt.Errorf("empty /llms.txt response")
	}

	files := make(map[string]string)
	for _, fullPath := range corpusPaths[1:] {
		sections, err := parseLLMSFull(resources[fullPath].body)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", fullPath, err)
		}
		for k, v := range sections {
			files[k] = v
		}
	}

	postCount := 0
	for k := range files {
		if strings.HasPrefix(k, "posts/") {
			postCount++
		}
	}

	return &Corpus{
		Files:        files,
		PromptIndex:  formatPromptIndexFromLLMS(llmsTxt),
		PostCount:    postCount,
		SectionCount: len(files),
		LoadedAt:     s.now(),
	}, nil
}

func normalizeCorpusKey(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if u, err := url.Parse(s); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		s = u.Path
	} else if idx := strings.IndexAny(s, "?#"); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/index.md")
	s = strings.TrimSuffix(s, "index.md")
	s = strings.TrimSuffix(s, "/")
	return s
}

func formatPromptIndexFromLLMS(llmsTxt string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(llmsTxt), "/index.md)", "/)")
	var b strings.Builder
	b.WriteString("\n## Master Content Index (Synced From danicat.dev/llms.txt)\n")
	b.WriteString("Call `read_corpus` with any listed URL or path (prepend `/pt-br/` or `/ja/` for Portuguese or Japanese editions) before quoting technical details, talks, videos, or projects.\n\n")
	b.WriteString(cleaned)
	return b.String()
}

func parseLLMSFull(raw string) (map[string]string, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	matches := llmsFullSectionRe.FindAllStringSubmatchIndex(normalized, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no sections found in llms-full.txt")
	}

	sections := make(map[string]string, len(matches))
	for i, m := range matches {
		start := m[0]
		end := len(normalized)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		rawURL := normalized[m[2]:m[3]]
		key := normalizeCorpusKey(rawURL)
		if key == "" {
			continue
		}
		section := strings.TrimSpace(normalized[start:end])
		if before, ok := strings.CutSuffix(section, "\n---"); ok {
			section = strings.TrimSpace(before)
		}
		sections[key] = section
	}
	if len(sections) == 0 {
		return nil, fmt.Errorf("no valid section keys parsed from llms-full.txt")
	}
	return sections, nil
}
