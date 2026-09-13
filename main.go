package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"myssg/view"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func main() {
	contents, err := listContents("content/*.md")
	if err != nil {
		log.Fatalf("failed to crawl the contents: %v", err)
	}

	p := parser.New()
	h := html.New()

	writer := writer{output: "./dist", prefix: "post"}

	for _, c := range contents {

		var buf bytes.Buffer
		doc := p.Parse(c.Data)
		if err := h.Render(&buf, c.Data, doc); err != nil {
			panic(err)
		}

		title := strings.ToLower(c.Meta.Title)
		page := view.Layout(c.Meta.Title, view.Content(buf.String()))

		if err := writer.writeHTML(title, page); err != nil {
			log.Fatalf("failed to write html file: %v\n", err)
		}
	}
}

type writer struct {
	output string
	prefix string
}

func (w writer) outputDir(name string) string {
	return filepath.Join(w.output, w.prefix, name)
}

func (w writer) createFile(name string) (*os.File, error) {
	dirPath := w.outputDir(name)

	if err := os.MkdirAll(dirPath, os.FileMode(0o755)); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	filename := filepath.Join(dirPath, "index.html")

	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return f, nil
}

func (w writer) writeHTML(name string, template templ.Component) error {
	if !filepath.IsLocal(name) {
		return fmt.Errorf("page name %q is not a valid output path", name)
	}

	f, err := w.createFile(name)
	if err != nil {
		return fmt.Errorf("failed to create file at destination path: %w", err)
	}
	defer f.Close()

	if err := template.Render(context.Background(), f); err != nil {
		return fmt.Errorf("failed to render index: %w", err)
	}

	return nil
}
