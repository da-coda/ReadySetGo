generate:
	go generate

run: generate
	go run main.go

templ_watch:
	bin/templ generate -watch
