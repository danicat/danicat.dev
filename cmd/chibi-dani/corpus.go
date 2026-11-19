package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const siteBaseURL = "https://danicat.dev"

// Corpus holds all published blog content in memory alongside a compact master index.
type Corpus struct {
	Files        map[string]string
	PromptIndex  string
	PostCount    int
	CodelabCount int
}

type frontmatter struct {
	Title       string   `yaml:"title"`
	Date        string   `yaml:"date"`
	Draft       bool     `yaml:"draft"`
	Slug        string   `yaml:"slug"`
	URL         string   `yaml:"url"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags"`
	Series      []string `yaml:"series"`
	SeriesOrder int      `yaml:"series_order"`
	Summary     string   `yaml:"summary"`
	Description string   `yaml:"description"`
	ExternalURL string   `yaml:"externalUrl"`
}

type postSummary struct {
	DirName     string
	Title       string
	Date        string
	Category    string
	Series      string
	SeriesOrder int
	Tags        []string
	Summary     string
	URLs        map[string]string
	Files       []string
}

type codelabSummary struct {
	Name        string
	Title       string
	Date        string
	ExternalURL string
	Summary     string
}

// LoadCorpus loads published posts, codelabs, about pages, talks, and videos into memory.
func LoadCorpus(rootDir string) (*Corpus, error) {
	root := strings.TrimSpace(rootDir)
	if root == "" {
		return nil, fmt.Errorf("corpus root directory must not be empty")
	}

	files := make(map[string]string)
	posts, err := loadPosts(root, files)
	if err != nil {
		return nil, err
	}
	codelabs, err := loadCodelabs(root)
	if err != nil {
		return nil, err
	}
	if err := loadAbout(root, files); err != nil {
		return nil, err
	}
	for _, name := range []string{"talks.json", "videos.json"} {
		data, err := os.ReadFile(filepath.Join(root, "data", name))
		if err != nil {
			return nil, fmt.Errorf("read data/%s: %w", name, err)
		}
		if !json.Valid(data) {
			return nil, fmt.Errorf("invalid JSON in data/%s", name)
		}
		files["data/"+name] = string(data)
	}

	return &Corpus{
		Files:        files,
		PostCount:    len(posts),
		CodelabCount: len(codelabs),
		PromptIndex:  formatPromptIndex(posts, codelabs),
	}, nil
}

func loadPosts(root string, files map[string]string) ([]postSummary, error) {
	postsRoot := filepath.Join(root, "content", "posts")
	entries, err := os.ReadDir(postsRoot)
	if err != nil {
		return nil, fmt.Errorf("read content/posts: %w", err)
	}

	var posts []postSummary
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}

		srcDir := filepath.Join(postsRoot, name)
		enPath := filepath.Join(srcDir, "index.md")
		data, err := os.ReadFile(enPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", enPath, err)
		}
		fm, err := parseFrontmatter(data)
		if err != nil {
			return nil, fmt.Errorf("parse frontmatter %s: %w", enPath, err)
		}
		if fm.Draft {
			continue
		}

		enKey := filepath.ToSlash(filepath.Join("posts", name, "index.md"))
		files[enKey] = string(data)

		slug := resolveSlug(name, fm)
		p := postSummary{
			DirName:     name,
			Title:       fm.Title,
			Date:        normalizeDate(fm.Date),
			Category:    firstString(fm.Categories),
			Series:      firstString(fm.Series),
			SeriesOrder: fm.SeriesOrder,
			Tags:        fm.Tags,
			Summary:     strings.Join(strings.Fields(cmp.Or(fm.Summary, fm.Description)), " "),
			URLs:        map[string]string{"en": canonicalPostURL("", slug)},
			Files:       []string{enKey},
		}

		for _, lang := range []string{"pt-br", "ja"} {
			filename := "index." + lang + ".md"
			lPath := filepath.Join(srcDir, filename)
			lData, err := os.ReadFile(lPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("read %s: %w", lPath, err)
			}
			lfm, err := parseFrontmatter(lData)
			if err != nil {
				return nil, fmt.Errorf("parse frontmatter %s: %w", lPath, err)
			}
			if lfm.Draft {
				continue
			}
			key := filepath.ToSlash(filepath.Join("posts", name, filename))
			files[key] = string(lData)
			p.Files = append(p.Files, key)
			p.URLs[lang] = canonicalPostURL(lang, resolveSlug(slug, lfm))
		}

		posts = append(posts, p)
	}

	slices.SortFunc(posts, func(a, b postSummary) int {
		if a.Date == b.Date {
			return strings.Compare(b.DirName, a.DirName)
		}
		return strings.Compare(b.Date, a.Date)
	})
	return posts, nil
}

func loadCodelabs(root string) ([]codelabSummary, error) {
	srcRoot := filepath.Join(root, "content", "codelabs")
	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read content/codelabs: %w", err)
	}

	var out []codelabSummary
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasPrefix(name, "_") || !strings.HasSuffix(name, ".md") {
			continue
		}
		fullPath := filepath.Join(srcRoot, name)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fullPath, err)
		}
		fm, err := parseFrontmatter(data)
		if err != nil {
			return nil, fmt.Errorf("parse frontmatter %s: %w", fullPath, err)
		}
		if fm.Draft || strings.Contains(name, ".pt-br.") || strings.Contains(name, ".ja.") {
			continue
		}
		out = append(out, codelabSummary{
			Name:        strings.TrimSuffix(name, ".md"),
			Title:       fm.Title,
			Date:        normalizeDate(fm.Date),
			ExternalURL: fm.ExternalURL,
			Summary:     strings.Join(strings.Fields(cmp.Or(fm.Summary, fm.Description)), " "),
		})
	}
	slices.SortFunc(out, func(a, b codelabSummary) int {
		if a.Date == b.Date {
			return strings.Compare(a.Name, b.Name)
		}
		return strings.Compare(b.Date, a.Date)
	})
	return out, nil
}

func loadAbout(root string, files map[string]string) error {
	srcDir := filepath.Join(root, "content", "about")
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		fm, err := parseFrontmatter(data)
		if err != nil {
			return fmt.Errorf("parse frontmatter %s: %w", path, err)
		}
		if fm.Draft {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		files["about/"+filepath.ToSlash(rel)] = string(data)
		return nil
	})
}

func parseFrontmatter(content []byte) (frontmatter, error) {
	var fm frontmatter
	trimmed := bytes.TrimSpace(content)
	if !bytes.HasPrefix(trimmed, []byte("---")) {
		return fm, nil
	}
	rest := bytes.TrimLeft(trimmed[3:], "\r\n")
	before, _, ok := bytes.Cut(rest, []byte("\n---"))
	if !ok {
		return fm, nil
	}
	if err := yaml.Unmarshal(before, &fm); err != nil {
		return fm, err
	}
	return fm, nil
}

func resolveSlug(fallback string, fm frontmatter) string {
	return strings.Trim(cmp.Or(fm.URL, fm.Slug, fallback), "/")
}

func canonicalPostURL(lang, slug string) string {
	slug = strings.Trim(slug, "/")
	if strings.HasPrefix(slug, "posts/") || strings.HasPrefix(slug, "pt-br/") || strings.HasPrefix(slug, "ja/") {
		return fmt.Sprintf("%s/%s/", siteBaseURL, slug)
	}
	if lang == "" {
		return fmt.Sprintf("%s/posts/%s/", siteBaseURL, slug)
	}
	return fmt.Sprintf("%s/%s/posts/%s/", siteBaseURL, lang, slug)
}

func normalizeDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}

func firstString(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[0])
}

func formatPromptIndex(posts []postSummary, codelabs []codelabSummary) string {
	var b strings.Builder
	b.WriteString("\n## Master Content Index (Auto-Generated From Published Source)\n")
	b.WriteString("Call `read_corpus` with any listed path before quoting technical details, talks, or projects.\n")
	b.WriteString("Core Pages:\n")
	b.WriteString("- About & Open Source Tools: https://danicat.dev/about/ (files: about/index.md, about/index.pt-br.md, about/index.ja.md)\n")
	b.WriteString("- Talks & Events: https://danicat.dev/events/ (file: data/talks.json)\n")
	b.WriteString("- Videos & Podcast Interviews: https://danicat.dev/videos/ (file: data/videos.json)\n")
	b.WriteString("- Codelabs & Workshops: https://danicat.dev/codelabs/\n")
	b.WriteString("- Agent Skills Catalog (KungFu): https://skills.danicat.dev/\n\n")

	fmt.Fprintf(&b, "### Published Blog Posts (%d total, newest first)\n", len(posts))
	for _, p := range posts {
		fmt.Fprintf(&b, "- **%s** (%s, %s)", p.Title, p.Date, p.Category)
		if p.Series != "" {
			fmt.Fprintf(&b, " [Series: %s #%d]", p.Series, p.SeriesOrder)
		}
		if len(p.Tags) > 0 {
			fmt.Fprintf(&b, " [tags: %s]", strings.Join(p.Tags, ", "))
		}
		b.WriteByte('\n')
		if p.Summary != "" {
			fmt.Fprintf(&b, "  Summary: %s\n", p.Summary)
		}
		fmt.Fprintf(&b, "  Canonical URLs: en=%s", p.URLs["en"])
		for _, lang := range []string{"pt-br", "ja"} {
			if u := p.URLs[lang]; u != "" {
				fmt.Fprintf(&b, " | %s=%s", lang, u)
			}
		}
		fmt.Fprintf(&b, "\n  Files: %s\n", strings.Join(p.Files, ", "))
	}

	fmt.Fprintf(&b, "\n### Published Codelabs (%d total)\n", len(codelabs))
	for _, c := range codelabs {
		fmt.Fprintf(&b, "- **%s** (%s): %s\n", c.Title, c.Date, c.ExternalURL)
		if c.Summary != "" {
			fmt.Fprintf(&b, "  Summary: %s\n", c.Summary)
		}
	}

	return b.String()
}
