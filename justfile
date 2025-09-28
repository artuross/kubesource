_list:
    just --list

build:
    CGO_ENABLED=0 go build -o dist/kubesource ./cmd/kubesource

format:
    golangci-lint fmt

run:
    go run ./cmd/kubesource
