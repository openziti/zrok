BIN := gowebdav
SRC := $(wildcard *.go) cmd/gowebdav/main.go

all: test cmd

cmd: ${BIN}

${BIN}: ${SRC}
	go build -o $@ ./cmd/gowebdav

test:
	go test -modfile=go_test.mod -v -short -cover ./...

api: .go/bin/godoc2md
	@sed '/^## API$$/,$$d' -i README.md
	@echo '## API' >> README.md
	@$< github.com/studio-b12/gowebdav | sed '/^$$/N;/^\n$$/D' |\
	sed '2d' |\
	sed 's/\/src\/github.com\/studio-b12\/gowebdav\//https:\/\/github.com\/studio-b12\/gowebdav\/blob\/master\//g' |\
	sed 's/\/src\/target\//https:\/\/github.com\/studio-b12\/gowebdav\/blob\/master\//g' |\
	sed 's/^#/##/g' >> README.md

check: .go/bin/gocyclo
	gofmt -w -s $(SRC)
	@echo
	.go/bin/gocyclo -over 15 .
	@echo
	go vet -modfile=go_test.mod ./...


.go/bin/godoc2md:
	@mkdir -p $(@D)
	@GOPATH="$(CURDIR)/.go" go install github.com/davecheney/godoc2md@latest

.go/bin/gocyclo:
	@mkdir -p $(@D)
	@GOPATH="$(CURDIR)/.go" go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

clean:
	@rm -f ${BIN}

.PHONY: all cmd clean test api check
