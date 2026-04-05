package cmd

import (
	"bytes"
	"strings"
	"text/template"

	gemara "github.com/gemaraproj/go-gemara"
)

var inlineFuncs = template.FuncMap{
	"inline": func(s string) string {
		return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	},
}

var capabilitiesTemplate = template.Must(template.New("capabilities").Funcs(inlineFuncs).Parse(`# {{ .Title }}

| ID | Title | Description |
|---|---|---|
{{- range .Capabilities }}
| {{ .Id }} | {{ .Title }} | {{ inline .Description }} |
{{- end }}
{{ if .Imports }}
## Imported Capabilities
{{ range .Imports }}
### From {{ .ReferenceId }}

| ID | Description |
|---|---|
{{- range .Entries }}
| {{ .ReferenceId }} | {{ .Remarks }} |
{{- end }}
{{ end }}
{{- end }}`))

var threatsTemplate = template.Must(template.New("threats").Funcs(inlineFuncs).Parse(`# {{ .Title }}

| ID | Title | Description |
|---|---|---|
{{- range .Threats }}
| {{ .Id }} | {{ .Title }} | {{ inline .Description }} |
{{- end }}
{{ if .Imports }}
## Imported Threats
{{ range .Imports }}
### From {{ .ReferenceId }}

| ID | Description |
|---|---|
{{- range .Entries }}
| {{ .ReferenceId }} | {{ .Remarks }} |
{{- end }}
{{ end }}
{{- end }}`))

func renderMarkdown(catalog *gemara.CapabilityCatalog) (string, error) {
	var buf bytes.Buffer
	if err := capabilitiesTemplate.Execute(&buf, catalog); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderThreatsMarkdown(catalog *gemara.ThreatCatalog) (string, error) {
	var buf bytes.Buffer
	if err := threatsTemplate.Execute(&buf, catalog); err != nil {
		return "", err
	}
	return buf.String(), nil
}
