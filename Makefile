# use bash
SHELL=/bin/bash

# set project VERSION if VERSION hasn't been passed from command line
ifndef $(value VERSION)
	VERSION=$(cat ./VERSION)
endif

# setup the -ldflags option for go build
LDFLAGS=-ldflags "-X main.Version=$(value VERSION)"

# build all executables
all:
	go build $(LDFLAGS) ./src/cmd/...

race:
	go build $(LDFLAGS) -race ./src/cmd/...

.PHONY: all clean install lint race

lint:
	reuse lint
	golangci-lint run
	
# move all executables to /usr/bin 
install:
	for CMD in `ls ./src/cmd`; do \
		install -Dm755 $$CMD $(DESTDIR)/usr/bin/$$CMD; \
	done

# remove build results
clean:
	for CMD in `ls ./src/cmd`; do \
		rm -f ./$$CMD; \
	done
