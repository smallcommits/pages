package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"myssg/web"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"go.yaml.in/yaml/v4"
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	mdParser    = parser.New(parser.WithExtensions(extension.GFMParser))
	mdRenderer  = html.New(html.WithExtensions(extension.GFMHTMLRenderer))
)

type frontmatter struct {
	Title       string `yaml:"title"`
	Date        any    `yaml:"date"` // unquoted YYYY-MM-DD decodes to time.Time, quoted to string
	Description string `yaml:"description"`
	Draft       bool   `yaml:"draft"`
}

// loadPosts reports every invalid file at once so the author can fix them in one pass.
func loadPosts(dir string, includeDrafts bool) ([]web.Post, error) {
	postsDir := filepath.Join(dir, "posts")
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}

	var posts []web.Post
	var errs []error
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".md" || strings.HasPrefix(name, "_") {
			continue
		}
		path := filepath.Join(postsDir, name)
		data, err := os.ReadFile(path)
		if err == nil {
			var p web.Post
			p, err = parsePost(strings.TrimSuffix(name, ".md"), data)
			if err == nil && (includeDrafts || !p.Draft) {
				posts = append(posts, p)
			}
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	slices.SortFunc(posts, func(a, b web.Post) int {
		return cmp.Or(b.Date.Compare(a.Date), cmp.Compare(a.Slug, b.Slug))
	})
	return posts, nil
}

func parsePost(slug string, data []byte) (web.Post, error) {
	if len(data) == 0 {
		return web.Post{}, errors.New("file is empty")
	}
	if !slugPattern.MatchString(slug) {
		return web.Post{}, fmt.Errorf("filename %q must be lowercase words joined by single hyphens", slug)
	}

	rest, ok := strings.CutPrefix(string(data), "---\n")
	if !ok {
		return web.Post{}, errors.New("file must start with --- frontmatter")
	}
	rawFM, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return web.Post{}, errors.New("frontmatter missing closing ---")
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(rawFM), &fm); err != nil {
		return web.Post{}, err
	}

	var errs []error
	if fm.Title == "" {
		errs = append(errs, errors.New("`title` is missing from frontmatter"))
	}
	var date time.Time
	switch d := fm.Date.(type) {
	case nil:
		errs = append(errs, errors.New("`date` is missing from frontmatter"))
	case time.Time:
		date = d
	default:
		var err error
		if date, err = time.Parse(time.DateOnly, fmt.Sprint(d)); err != nil {
			errs = append(errs, fmt.Errorf("`date` %v must be YYYY-MM-DD", d))
		}
	}
	if len(errs) > 0 {
		return web.Post{}, errors.Join(errs...)
	}

	src := []byte(body)
	var buf bytes.Buffer
	if err := mdRenderer.Render(&buf, src, mdParser.Parse(src)); err != nil {
		return web.Post{}, err
	}

	return web.Post{
		Slug:        slug,
		Title:       fm.Title,
		Date:        date,
		Description: fm.Description,
		Draft:       fm.Draft,
		Body:        buf.String(),
	}, nil
}
