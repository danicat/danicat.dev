{{- $summary := (.Params.description | default (.Summary | plainify | htmlUnescape)) | replaceRE "\\s+" " " -}}
{{- $site := .Site -}}
{{- $statusLabels := dict "evergreen" "Evergreen" "historical" "Historical" "superseded" "Superseded" "deprecated" "Deprecated" -}}
{{- $status := lower (.Params.status | default "") -}}
{{- $supersededUrls := slice -}}
{{- $supersededLinks := slice -}}
{{- $supersededIsSlice := false -}}
{{- with .Params.superseded_by -}}
  {{- $refs := slice -}}
  {{- if reflect.IsSlice . -}}
    {{- $supersededIsSlice = true -}}
    {{- $refs = . -}}
  {{- else -}}
    {{- $refs = slice . -}}
  {{- end -}}
  {{- range $refs -}}
    {{- $ref := . -}}
    {{- with $site.GetPage $ref -}}
      {{- if .Permalink -}}
        {{- $supersededUrls = $supersededUrls | append .Permalink -}}
        {{- $supersededLinks = $supersededLinks | append (printf "[%s](%sindex.md) ([HTML](%s))" (.Title | plainify | htmlUnescape) .Permalink .Permalink) -}}
      {{- else if .Params.externalUrl -}}
        {{- $supersededUrls = $supersededUrls | append .Params.externalUrl -}}
        {{- $supersededLinks = $supersededLinks | append (printf "[%s](%s)" (.Title | plainify | htmlUnescape) .Params.externalUrl) -}}
      {{- else -}}
        {{- $supersededUrls = $supersededUrls | append $ref -}}
        {{- $supersededLinks = $supersededLinks | append (.Title | plainify | htmlUnescape) -}}
      {{- end -}}
    {{- else -}}
      {{- $supersededUrls = $supersededUrls | append $ref -}}
      {{- $supersededLinks = $supersededLinks | append $ref -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
---
title: {{ .Title | plainify | htmlUnescape | jsonify }}
{{- if not .Date.IsZero }}
date: {{ .Date.Format "2006-01-02" }}
{{- end }}
{{- if not .Lastmod.IsZero }}
lastmod: {{ .Lastmod.Format "2006-01-02" }}
{{- end }}
{{- with $summary }}
description: {{ . | jsonify }}
{{- end }}
{{- with .Params.categories }}
categories: {{ . | jsonify }}
{{- end }}
{{- with .Params.tags }}
tags: {{ . | jsonify }}
{{- end }}
{{- with .Params.series }}
series: {{ . | jsonify }}
{{- end }}
{{- with .Params.series_order }}
series_order: {{ . }}
{{- end }}
{{- with .Params.proficiencyLevel }}
proficiencyLevel: {{ . | jsonify }}
{{- end }}
{{- with .Params.dependencies }}
dependencies: {{ . | jsonify }}
{{- end }}
{{- with $status }}
status: {{ . | jsonify }}
{{- end }}
{{- if gt (len $supersededUrls) 0 }}
superseded_by: {{ if $supersededIsSlice }}{{ $supersededUrls | jsonify (dict "noHTMLEscape" true) }}{{ else }}{{ index $supersededUrls 0 | jsonify (dict "noHTMLEscape" true) }}{{ end }}
{{- end }}
{{- with .Params.status_note }}
status_note: {{ . | replaceRE "\\s+" " " | strings.TrimSpace | jsonify (dict "noHTMLEscape" true) }}
{{- end }}
language: {{ .Language.LanguageCode | jsonify }}
{{- with .Translations }}
translations:
{{- range . }}
  - lang: {{ .Language.LanguageCode | jsonify }}
    url: {{ .Permalink }}
    markdown_url: {{ .Permalink }}index.md
{{- end }}
{{- end }}
url: {{ .Permalink }}
---

# {{ .Title | plainify | htmlUnescape }}{{ if eq $status "deprecated" }} [DEPRECATED]{{ else if eq $status "superseded" }} [SUPERSEDED]{{ end }}

> For the complete documentation index, see [llms.txt]({{ "llms.txt" | absLangURL }}) or the full site corpus at [llms-full.txt]({{ "llms-full.txt" | absLangURL }}).

- **URL**: {{ .Permalink }}
- **Markdown URL**: {{ .Permalink }}index.md
- **Author**: {{ .Site.Params.Author.name | default "Daniela Petruzalek" }}
{{- with .Params.role }}
- **Role**: {{ . }}
{{- end }}
{{- with .Params.badges }}
- **Highlights**: {{ delimit . " · " }}
{{- end }}
{{- if not .Date.IsZero }}
- **Date**: {{ .Date.Format "2006-01-02" }}
{{- end }}
{{- with index $statusLabels $status }}
- **Status**: {{ . }}
{{- end }}
{{- if gt (len $supersededLinks) 0 }}
- **Superseded by**: {{ delimit $supersededLinks ", " }}
{{- end }}
{{- with .Params.status_note }}
- **Status Note**: {{ . | replaceRE "\\s+" " " | strings.TrimSpace }}
{{- end }}
{{- with .Params.categories }}
- **Categories**: {{ delimit . ", " }}
{{- end }}
{{- with .Params.tags }}
- **Tags**: {{ delimit . ", " }}
{{- end }}
{{- with .Params.series }}
{{- $seriesName := index . 0 }}
- **Series**: {{ $seriesName }}{{ with $.Params.series_order }} (Part {{ . }}){{ end }}
{{- $seriesTax := or (index $.Site.Taxonomies.series ($seriesName | lower)) (index $.Site.Taxonomies.series ($seriesName | urlize)) }}
{{- with $seriesTax }}
{{- range $post := .Pages.ByParam "series_order" }}
  - {{ with $post.Params.series_order }}Part {{ . }}: {{ end }}[{{ $post.Title | plainify | htmlUnescape }}]({{ $post.Permalink }}index.md) ([HTML]({{ $post.Permalink }})){{ if eq $post.Permalink $.Permalink }} *(this article)*{{ end }}
{{- end }}
{{- end }}
{{- end }}
{{- with .Translations }}
- **Translations**: {{ range $i, $t := . }}{{ if $i }}, {{ end }}[{{ $t.Language.LanguageName }} ({{ $t.Language.LanguageCode }})]({{ $t.Permalink }}index.md) ([HTML]({{ $t.Permalink }})){{ end }}
{{- end }}
{{- with $summary }}
- **Summary**: {{ . }}
{{- end }}

---

{{- $baseOrigin := strings.TrimSuffix "/" site.BaseURL -}}
{{- $assetBase := .Permalink | replaceRE "^(https?://[^/]+)/(?:pt-br|ja)/" "${1}/" -}}
{{- $body := .RawContent -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*mermaid\s*[%>]\}\}` "```mermaid" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*/mermaid\s*[%>]\}\}` "```" -}}
{{- range $match := findRE `\{\{[<%]\s*(?:rel)?ref\s+"[^"]+"\s*[%>]\}\}` $body -}}
  {{- $rawRef := replaceRE `^\{\{[<%]\s*(?:rel)?ref\s+"/?([^"#]+)(?:#[^"]*)?"\s*[%>]\}\}$` "${1}" $match -}}
  {{- $cleanRef := $rawRef | replaceRE `(?:/index(?:\.[a-z-]+)?\.md|\.md|/)$` "" -}}
  {{- $frag := "" -}}
  {{- if in $match "#" -}}
    {{- $frag = replaceRE `^\{\{[<%]\s*(?:rel)?ref\s+"[^"#]+(#[^"]+)"\s*[%>]\}\}$` "${1}" $match -}}
  {{- end -}}
  {{- with $.Site.GetPage (printf "/%s" $cleanRef) -}}
    {{- $body = replace $body $match (printf "%s%s" .Permalink $frag) -}}
  {{- end -}}
{{- end -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?id=["']([A-Za-z0-9_-]+)["'][^>]*?title=["']([^"']+)["'][^>]*?[%>]\}\}` "[YouTube Video: ${2}](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?title=["']([^"']+)["'][^>]*?id=["']([A-Za-z0-9_-]+)["'][^>]*?[%>]\}\}` "[YouTube Video: ${1}](https://www.youtube.com/watch?v=${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?id=["']?([A-Za-z0-9_-]+)["']?[^>]*?[%>]\}\}` "[YouTube Video](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+["']?([A-Za-z0-9_-]+)["']?\s*[%>]\}\}` "[YouTube Video](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*(?:x|twitter)\s+user=["']([^"']+)["']\s+id=["']([^"']+)["']\s*[%>]\}\}` "[Post on X](https://x.com/${1}/status/${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*gist\s+([^\s"']+)\s+([^\s"']+)(?:\s+["']?([^"'\s>}]+)["']?)?\s*[%>]\}\}` "[GitHub Gist: ${1}/${2}](https://gist.github.com/${1}/${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*github\s+user=["']([^"']+)["']\s+repo=["']([^"']+)["']\s+ref=["']([^"']+)["']\s+path=["']([^"']+)["'](?:\s+lang=["'][^"']+["'])?\s+start=["']([0-9]+)["']\s+end=["']([0-9]+)["']\s*[%>]\}\}` "[View on GitHub: ${1}/${2}/${4} (lines ${5}-${6})](https://github.com/${1}/${2}/blob/${3}/${4}#L${5}-L${6})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*github\s+user=["']([^"']+)["']\s+repo=["']([^"']+)["']\s+path=["']([^"']+)["'](?:\s+lang=["'][^"']+["'])?\s+start=["']([0-9]+)["']\s+end=["']([0-9]+)["']\s*[%>]\}\}` "[View on GitHub: ${1}/${2}/${3} (lines ${4}-${5})](https://github.com/${1}/${2}/blob/main/${3}#L${4}-L${5})" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*lead\s*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/lead\s*[%>]\}\}` "${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*alert[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/alert\s*[%>]\}\}` "> ${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*warning[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/warning\s*[%>]\}\}` "> **Warning:** ${1}" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*figure\s+[^>]*src=["']([^"']+)["'][^>]*[%>]\}\}` "![Figure](${1})" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*badge[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/badge\s*[%>]\}\}` "${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*button[^\}]*href=["']([^"']+)["'][^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/button\s*[%>]\}\}` "[${2}](${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*katex\s*[%>]\}\}\n?` "" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*/?[a-zA-Z0-9_-]+[^>]*[%>]\}\}` "" -}}
{{- $body = $body | replaceRE `\[([^\]]+)\]\((/[^)\s]*)\)` (printf "[${1}](%s${2})" $baseOrigin) -}}
{{- $body = $body | replaceRE `\[([^\]]+)\]\(([a-zA-Z0-9_.-]+\.[a-zA-Z0-9]+)\)` (printf "[${1}](%s${2})" $assetBase) -}}

{{ $body }}
{{- with .Params.speaking }}

## {{ .heading }}
{{- with .subtitle }}

{{ . }}
{{- end }}

{{- range .cards }}
{{- $cardUrl := .url | absURL }}
- **{{ .icon }} [{{ .title }}]({{ $cardUrl }})** ({{ .cta }}): {{ .description }}
{{- end }}
{{- end }}
{{- with .Params.projects }}

## {{ .heading }}
{{- with .subtitle }}

{{ . }}
{{- end }}

{{- range .items }}
{{- $itemUrl := .url | absURL }}
- **[{{ .name }}]({{ $itemUrl }})** ({{ .badge }}): {{ .description }}
{{- end }}
{{- end }}
