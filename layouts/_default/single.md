# {{ .Title }}

- **URL**: {{ .Permalink }}
- **Author**: {{ .Site.Params.Author.name | default "Daniela Petruzalek" }}
{{- if not .Date.IsZero }}
- **Date**: {{ .Date.Format "2006-01-02" }}
{{- end }}
{{- with .Params.categories }}
- **Categories**: {{ delimit . ", " }}
{{- end }}
{{- with (.Params.description | default (.Summary | plainify | htmlUnescape)) }}
- **Summary**: {{ . | replaceRE "\\s+" " " }}
{{- end }}

---

{{ .RawContent }}
