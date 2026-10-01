package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"myssg/web"
)

func TestParsePost(t *testing.T) {
	if q, err := parsePost("quoted", []byte("---\ntitle: Q\ndate: \"2026-10-01\"\n---\n")); err != nil || q.Date.Day() != 1 {
		t.Fatalf("quoted date: %v %v", q.Date, err)
	}

	p, err := parsePost("first-post", []byte("---\ntitle: Hi\ndate: 2026-10-01\ndescription: About\n---\nbody *text*\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if p.Slug != "first-post" || p.Title != "Hi" || !p.Date.Equal(want) || p.Description != "About" || p.Draft {
		t.Fatalf("unexpected post %+v", p)
	}
	if p.Body != "<p>body <em>text</em></p>\n" {
		t.Fatalf("body = %q", p.Body)
	}

	for name, tc := range map[string]struct{ slug, in, wantErr string }{
		"empty file":         {"ok", "", "empty"},
		"no frontmatter":     {"ok", "# just markdown\n", "must start with ---"},
		"unclosed":           {"ok", "---\ntitle: Hi\ndate: 2026-10-01\n\n# body", "closing ---"},
		"missing title":      {"ok", "---\ndate: 2026-10-01\n---\nbody", "`title`"},
		"missing date":       {"ok", "---\ntitle: Hi\n---\nbody", "`date` is missing"},
		"invalid date":       {"ok", "---\ntitle: Hi\ndate: 01/10/2026\n---\nbody", "YYYY-MM-DD"},
		"uppercase slug":     {"Hello", "---\ntitle: Hi\ndate: 2026-10-01\n---\nbody", "filename"},
		"double hyphen slug": {"a--b", "---\ntitle: Hi\ndate: 2026-10-01\n---\nbody", "filename"},
	} {
		if _, err := parsePost(tc.slug, []byte(tc.in)); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: err = %v, want containing %q", name, err, tc.wantErr)
		}
	}
}

func writePosts(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "posts"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, "posts", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

var testPosts = map[string]string{
	"older.md":     "---\ntitle: Older\ndate: 2026-09-01\n---\nold\n",
	"newer.md":     "---\ntitle: Newer\ndate: 2026-10-01\n---\n| a | b |\n|---|---|\n| 1 | 2 |\n",
	"wip.md":       "---\ntitle: WIP\ndate: 2026-10-02\ndraft: true\n---\nunfinished\n",
	"_template.md": "not a post",
}

func TestLoadPostsDrafts(t *testing.T) {
	dir := writePosts(t, testPosts)

	posts, err := loadPosts(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := slugs(posts); got != "newer,older" {
		t.Fatalf("without drafts got %s", got)
	}

	posts, err = loadPosts(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := slugs(posts); got != "wip,newer,older" {
		t.Fatalf("with drafts got %s", got)
	}
}

func TestLoadPostsReportsEveryBadFile(t *testing.T) {
	dir := writePosts(t, map[string]string{
		"good.md":     "---\ntitle: Good\ndate: 2026-10-01\n---\nok\n",
		"no-title.md": "---\ndate: 2026-10-01\n---\nx\n",
		"empty.md":    "",
	})
	_, err := loadPosts(dir, false)
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, f := range []string{"no-title.md", "empty.md"} {
		if !strings.Contains(err.Error(), f) {
			t.Errorf("error does not mention %s: %v", f, err)
		}
	}
}

func TestBuild(t *testing.T) {
	dir := writePosts(t, testPosts)
	out := filepath.Join(t.TempDir(), "dist")
	stale := filepath.Join(out, "post", "index.html")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	posts, err := loadPosts(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := build(posts, "Test Blog", out); err != nil {
		t.Fatal(err)
	}

	index := read(t, filepath.Join(out, "index.html"))
	for _, want := range []string{`href="/posts/newer/"`, `href="/posts/older/"`, "Test Blog"} {
		if !strings.Contains(index, want) {
			t.Errorf("index.html missing %s", want)
		}
	}
	if strings.Contains(index, "/posts/wip/") {
		t.Error("index.html links the draft")
	}
	if post := read(t, filepath.Join(out, "posts", "newer", "index.html")); !strings.Contains(post, "<td>1</td>") {
		t.Errorf("post page missing rendered table:\n%s", post)
	}
	for _, path := range []string{"404.html", "static/style.css"} {
		read(t, filepath.Join(out, path))
	}
	for _, path := range []string{"posts/wip", "post"} {
		if _, err := os.Stat(filepath.Join(out, path)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist, stat err = %v", path, err)
		}
	}
}

func slugs(posts []web.Post) string {
	s := make([]string, len(posts))
	for i, p := range posts {
		s[i] = p.Slug
	}
	return strings.Join(s, ",")
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
