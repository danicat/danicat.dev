{{- $summary := (.Params.description | default (.Summary | plainify | htmlUnescape)) | replaceRE "\\s+" " " -}}
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

# {{ .Title | plainify | htmlUnescape }}

- **URL**: {{ .Permalink }}
- **Author**: {{ .Site.Params.Author.name | default "Daniela Petruzalek" }}
{{- if not .Date.IsZero }}
- **Date**: {{ .Date.Format "2006-01-02" }}
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
  - {{ with $post.Params.series_order }}Part {{ . }}: {{ end }}[{{ $post.Title | plainify | htmlUnescape }}]({{ $post.Permalink }}) ([Markdown]({{ $post.Permalink }}index.md)){{ if eq $post.Permalink $.Permalink }} *(this article)*{{ end }}
{{- end }}
{{- end }}
{{- end }}
{{- with .Translations }}
- **Translations**: {{ range $i, $t := . }}{{ if $i }}, {{ end }}[{{ $t.Language.LanguageName }} ({{ $t.Language.LanguageCode }})]({{ $t.Permalink }}) ([Markdown]({{ $t.Permalink }}index.md)){{ end }}
{{- end }}
{{- with $summary }}
- **Summary**: {{ . }}
{{- end }}

---

{{- $body := .RawContent -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*mermaid\s*[%>]\}\}` "```mermaid" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*/mermaid\s*[%>]\}\}` "```" -}}
{{- $body = $body | replaceRE `(\{\{[<%]\s*(?:rel)?ref\s+"[^"#]*?)(?:/index\.md|\.md)(#[^"]*)?("\s*[%>]\}\})` "${1}${2}${3}" -}}
{{- $body = $body | replaceRE `(\{\{[<%]\s*(?:rel)?ref\s+"[^"#]*?)/(#[^"]*)?("\s*[%>]\}\})` "${1}${2}${3}" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*(?:rel)?ref\s+"/?([^"#]+)(#[^"]+)?"\s*[%>]\}\}` (printf "https://danicat.dev%s/${1}/${2}" site.LanguagePrefix) -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?id=["']([A-Za-z0-9_-]+)["'][^>]*?title=["']([^"']+)["'][^>]*?[%>]\}\}` "[YouTube Video: ${2}](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?title=["']([^"']+)["'][^>]*?id=["']([A-Za-z0-9_-]+)["'][^>]*?[%>]\}\}` "[YouTube Video: ${1}](https://www.youtube.com/watch?v=${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+[^>]*?id=["']?([A-Za-z0-9_-]+)["']?[^>]*?[%>]\}\}` "[YouTube Video](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*youtube\s+["']?([A-Za-z0-9_-]+)["']?\s*[%>]\}\}` "[YouTube Video](https://www.youtube.com/watch?v=${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*(?:x|twitter)\s+user=["']([^"']+)["']\s+id=["']([^"']+)["']\s*[%>]\}\}` "[Post on X](https://x.com/${1}/status/${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*gist\s+([^\s"']+)\s+([^\s"']+)(?:\s+["']?([^"'\s>}]+)["']?)?\s*[%>]\}\}` "[GitHub Gist: ${1}/${2}](https://gist.github.com/${1}/${2})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*github\s+user=["']([^"']+)["']\s+repo=["']([^"']+)["']\s+path=["']([^"']+)["'](?:\s+lang=["'][^"']+["'])?\s+start=["']([0-9]+)["']\s+end=["']([0-9]+)["']\s*[%>]\}\}` "[View on GitHub: ${1}/${2}/${3} (lines ${4}-${5})](https://github.com/${1}/${2}/blob/main/${3}#L${4}-L${5})" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*lead\s*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/lead\s*[%>]\}\}` "${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*alert[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/alert\s*[%>]\}\}` "> ${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*warning[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/warning\s*[%>]\}\}` "> **Warning:** ${1}" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*figure\s+[^>]*src=["']([^"']+)["'][^>]*[%>]\}\}` "![Figure](${1})" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*badge[^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/badge\s*[%>]\}\}` "${1}" -}}
{{- $body = $body | replaceRE `(?s)\{\{[<%]\s*button[^\}]*href=["']([^"']+)["'][^\}]*[%>]\}\}\s*(.*?)\s*\{\{[<%]\s*/button\s*[%>]\}\}` "[${2}](${1})" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*katex\s*[%>]\}\}\n?` "" -}}
{{- $body = $body | replaceRE `\{\{[<%]\s*/?[a-zA-Z0-9_-]+[^>]*[%>]\}\}` "" -}}

{{ $body }}
