package tracker

import (
	_ "embed"
	"html/template"

	"github.com/adamdecaf/community-commits/internal/source"
)

var (

	//go:embed templates/index.html.tmpl
	indexTemplateData string
	IndexTemplate     = template.Must(template.New("index").Parse(indexTemplateData))
)

type PushEventsTemplate struct {
	Date    string
	Commits []source.PushEvent
}

type IndexTemplateData struct {
	PushEvents []PushEventsTemplate
}
