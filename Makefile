# use bash
SHELL=/usr/bin/bash

# set project VERSION to last tag name. If no tag exists, set it to v0.0.0
$(eval TAGS=$(shell git rev-list --tags))
ifdef TAGS
	VERSION=$(shell git describe --tags --abbrev=0)
else
	VERSION=v0.0.0	
endif

.PHONY: all clean install lint race release

# setup the -ldflags option for go build
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# build all executables
all:
	go build $(LDFLAGS) ./src/cmd/...

race:
	go build $(LDFLAGS) -race ./src/cmd/...

lint:
	reuse lint
	golangci-lint run
	
# move all executables to /usr/bin 
install:
	@for CMD in `ls ./src/cmd`; do \
		install -Dm755 $$CMD $(DESTDIR)/usr/bin/$$CMD; \
	done

# remove build results
clean:
	@for CMD in `ls ./src/cmd`; do \
		rm -f ./$$CMD; \
	done

# (1) adjust version in PKGBUILD and in man documentation to RELEASE, commit
#     and push changes
# (2) create an annotated tag with name RELEASE
# syntax: make release RELEASE=vX.Y.Z
release:
	@if ! [ -z $(RELEASE) ]; then \
		REL=$(RELEASE); \
		sed -i -e "s/pkgver=.*/pkgver=$${REL#v}/" ./pkg/AUR/PKGBUILD; \
		git commit -a -s -m "release $(RELEASE)"; \
		git push; \
		git tag -a $(RELEASE) -m "release $(RELEASE)"; \
		git push origin $(RELEASE); \
	fi