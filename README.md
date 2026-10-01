# my-ssg

A small static site generator for a Markdown blog, deployed to Cloudflare Pages.

## Write a post

1. Copy `content/posts/_template.md` to `content/posts/<slug>.md`. The slug is
   lowercase words joined by hyphens, and it becomes the URL `/posts/<slug>/`.
2. Fill in the frontmatter. `title` and `date` (`YYYY-MM-DD`) are required.
   `description` is optional. `draft: true` keeps the post out of the build.
3. Write the body in Markdown. Tables, strikethrough, and task lists work.

Files that start with `_` are ignored.

## Build and preview

```sh
just build   # writes dist/
just serve   # builds, then serves dist/ at http://localhost:8080
just test
```

To preview drafts, run `go run . -drafts`. Run `go run . -h` to list the other
flags (`-content`, `-out`, `-site-title`).

## Deploy

Upload `dist/` to Cloudflare Pages:

```sh
just build
npx wrangler pages deploy dist --project-name=my-blog
```
