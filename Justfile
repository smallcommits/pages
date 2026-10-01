gen:
  go tool templ generate

build: gen
  go run .

test: gen
  go test ./...

serve: build
  python3 -m http.server -d dist 8080
