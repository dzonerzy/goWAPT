# GOWPT - Makefile
# global variables
GO=$(shell which go)
OUTFILE=gowapt
SOURCEDIR=src
INSTALLDIR=/usr/local/bin/


# Do not touch these!
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
DEPS = github.com/nsf/termbox-go golang.org/x/net/html github.com/robertkrimen/otto github.com/elazarl/goproxy

gowpt:
	$(info Remember to set GOPATH!)
	$(info Downloading dependencies $(DEPS))
	$(foreach var,$(DEPS),$(GO) get $(var);)
	$(GO) build -ldflags="-s -w" -o $(OUTFILE) $(SOURCES)

install:
	install -m 755 $(OUTFILE) $(INSTALLDIR)
	rm -f $(OUTFILE)

.PHONY: clean

clean:
	rm -f $(OUTFILE)
