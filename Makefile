# alloydb build rules.
ARCH="`uname -s`"

LINUX="Linux"
MAC="Darwin"

GO=godep go

LDFLAGS += -X "github.com/Dong-Chan/alloydb/util/printer.TiDBBuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/Dong-Chan/alloydb/util/printer.TiDBGitHash=$(shell git rev-parse HEAD)"

.PHONY: deps all build install parser clean todo test tidbtest mysqltest gotest interpreter

all: godep parser build test

godep:
	go get github.com/tools/godep

build:
	$(GO) build

install:
	$(GO) install ./...

parser:
	go get github.com/qiuyesuifeng/goyacc
	go get github.com/qiuyesuifeng/golex
	a=`mktemp temp.XXXXXX`; \
	goyacc -o /dev/null -xegen $$a parser/parser.y; \
	goyacc -o parser/parser.go -xe $$a parser/parser.y; \
	rm -f $$a; \
	rm -f y.output

	@if [ $(ARCH) = $(LINUX) ]; \
	then \
		sed -i -e 's|//line.*||' -e 's/yyEofCode/yyEOFCode/' parser/parser.go; \
	elif [ $(ARCH) = $(MAC) ]; \
	then \
		sed -i "" 's|//line.*||' parser/parser.go; \
		sed -i "" 's/yyEofCode/yyEOFCode/' parser/parser.go; \
	fi

	golex -o parser/scanner.go parser/scanner.l

deps:
	go list -f '{{range .Deps}}{{printf "%s\n" .}}{{end}}{{range .TestImports}}{{printf "%s\n" .}}{{end}}' ./... | \
		sort | uniq | grep -E '[^/]+\.[^/]+/' |grep -v "Dong-Chan/alloydb" | \
		awk 'BEGIN{ print "#!/bin/bash" }{ printf("go get -u %s\n", $$1) }' > deps.sh
	chmod +x deps.sh
	bash deps.sh

clean:
	$(GO) clean -i ./...
	rm -rf *.out
	rm -f deps.sh

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* */*.go parser/scanner.l parser/parser.y || true
	@grep -n TODO */*.go parser/scanner.l parser/parser.y alloydb-test/testdata.ql || true
	@grep -n BUG */*.go parser/scanner.l parser/parser.y || true
	@grep -n println */*.go parser/scanner.l parser/parser.y || true

test: gotest 

gotest:
	$(GO) test -cover ./...

race:
	$(GO) test --race -cover ./...

interpreter:
	@cd interpreter && $(GO) build -ldflags '$(LDFLAGS)'
