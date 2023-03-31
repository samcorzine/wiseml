#PHONY: run
build:
	nix-shell -p go --run "go build -o bin/main"

run: build
	./bin/main