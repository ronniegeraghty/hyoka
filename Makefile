.PHONY: build run test lint clean

build:
	cd tool && go build -o ../bin/azsdk-prompt-eval ./cmd/azsdk-prompt-eval

run:
	cd tool && go run ./cmd/azsdk-prompt-eval $(ARGS)

test:
	cd tool && go test ./...

lint:
	cd tool && go vet ./...

clean:
	rm -rf bin/
