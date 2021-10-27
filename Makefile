INSTALL=/usr/local/bin/

all:
	go build .

install:
	GOBIN=$(INSTALL) go install .