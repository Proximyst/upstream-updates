SRC = $(wildcard *.go)

.PHONY: build
build: $(SRC) go.mod go.sum
	go build -ldflags "-extldflags \"-static -fno-PIC\"" -buildmode pie -tags 'osusergo netgo static_build' .
