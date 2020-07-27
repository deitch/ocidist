PHONY: build
DISTDIR = dist
BIN = $(DISTDIR)/ocidist

export GO111MODULE=on

$(DISTDIR):
	mkdir -p $@

build: $(BIN)
$(BIN): $(DISTDIR)
	go build -o $@ .

clean:
	rm -rf $(BIN)

install:
	go install .
