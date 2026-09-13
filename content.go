package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

type Meta struct {
	Title string `yaml:"title"`
}

func (m Meta) Validate() []string {
	errs := []string{}

	if m.Title == "" {
		errs = append(errs, "`title` is missing from frontmatter")
	}

	return errs
}

type Content struct {
	Meta Meta
	Data []byte
}

func listContents(pattern string) ([]Content, error) {
	entries, err := filepath.Glob(pattern)
	if errors.Is(err, filepath.ErrBadPattern) {
		return nil, err
	}

	contents := make([]Content, 0, len(entries))

	for _, path := range entries {

		filename := filepath.Base(path)

		// skip not markdown file and file start with underscore
		if !strings.HasSuffix(filename, ".md") || strings.HasPrefix(filename, "_") {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("[warn] file %s cannot be read: %v\n", path, err)
			continue
		}
		if len(content) == 0 {
			fmt.Printf("[warn] file %s is empty\n", path)
			continue
		}

		data, err := parseContent(content)
		if err != nil {
			log.Fatalf("failed to parse content in file %s: %v", path, err)
		}
		contents = append(contents, data)
	}

	return contents, err
}

func parseContent(data []byte) (Content, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return Content{}, errors.New("frontmatter with `title` is missing")
	}

	fm, body, found := strings.Cut(
		strings.TrimPrefix(s, "---\n"),
		"\n---\n",
	)
	if !found {
		return Content{}, fmt.Errorf("frontmatter missing closing ---")
	}

	var meta Meta
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return Content{}, err
	}

	errs := meta.Validate()
	if len(errs) > 0 {
		return Content{}, errors.New(
			strings.Join(errs, ", "),
		)
	}

	return Content{
		Meta: meta,
		Data: []byte(body),
	}, nil
}
