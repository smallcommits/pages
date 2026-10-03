package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"myssg/web"

	"github.com/a-h/templ"
)

func main() {
	contentDir := flag.String("content", "./content", "directory containing commits/ and images/")
	out := flag.String("out", "./dist", "output directory, removed and rebuilt on every run")
	drafts := flag.Bool("drafts", false, "include posts marked draft: true")
	siteTitle := flag.String("site-title", "Small Commits", "site title shown in the header")
	flag.Parse()

	// build deletes out, so refuse anything that could reach outside the repo or wipe it.
	if !filepath.IsLocal(*out) || filepath.Clean(*out) == "." {
		log.Fatalf("-out %q must be a subdirectory of the current directory", *out)
	}

	posts, err := loadPosts(*contentDir, *drafts)
	if err != nil {
		log.Fatalf("invalid content:\n%v", err)
	}
	if err := build(posts, *siteTitle, *contentDir, *out); err != nil {
		log.Fatalf("build failed: %v", err)
	}
	fmt.Printf("built %d posts into %s\n", len(posts), *out)
}

type page struct {
	path string
	page templ.Component
}

func build(posts []web.Post, siteTitle, contentDir, out string) error {
	if err := os.RemoveAll(out); err != nil {
		return err
	}

	pages := []page{
		{"index.html", web.Index(siteTitle, posts)},
		{"404.html", web.NotFound(siteTitle)},
	}
	for _, p := range posts {
		pages = append(pages, page{filepath.Join("commits", p.Slug, "index.html"), web.PostPage(siteTitle, p)})
	}

	for _, p := range pages {
		if err := writePage(filepath.Join(out, p.path), p.page); err != nil {
			return fmt.Errorf("%s: %w", p.path, err)
		}
	}

	static, err := fs.Sub(web.Static, "static")
	if err != nil {
		return err
	}
	if err := os.CopyFS(filepath.Join(out, "static"), static); err != nil {
		return err
	}

	images := filepath.Join(contentDir, "images")
	if _, err := os.Stat(images); os.IsNotExist(err) {
		return nil
	}
	return os.CopyFS(filepath.Join(out, "images"), os.DirFS(images))
}

func writePage(path string, c templ.Component) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := c.Render(context.Background(), f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
