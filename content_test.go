package main

import "testing"

func TestContentParser(t *testing.T) {
	c, err := parseContent([]byte("---\ntitle: Hi\n---\nbody text"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Meta.Title != "Hi" || string(c.Data) != "body text" {
		t.Fatalf("got %q %q", c.Meta.Title, c.Data)
	}

	c, err = parseContent([]byte("---\ntitle: Hi\n---"))
	if err == nil {
		t.Fatal("expected an error for unterminated content")
	}

	for name, in := range map[string]string{
		"no frontmatter": "# just markdown\nno frontmatter",
		"no title":       "---\nfoo: bar\n---\nbody",
		"unclosed fence": "---\ntitle: Hi\n\n# body",
	} {
		if got, err := parseContent([]byte(in)); err == nil {
			t.Fatalf("%s: expected an error, got %q", name, got.Data)
		}
	}
}
