gen:
  go tool templ generate

build: gen
  go run .
