GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)
.PHONY: statik

build: bin vendor
	go build \
		-o bin/drugstore \
		-ldflags "-X github.com/factorysh/drugstore/version.version=$(GIT_VERSION)" \
		.

bin:
	mkdir -p bin
	chmod 777 bin

vendor:
	dep ensure

clean:
	rm -rf bin vendor

up:
	docker-compose up -d
	sleep 3

down:
	docker-compose down

psql: up
	PGPASSWORD=toto psql -h localhost -U drugstore drugstore

test:
	go test -v github.com/factorysh/drugstore/store
	go test -v github.com/factorysh/drugstore/rpc
	go test -v github.com/factorysh/drugstore/rest

statik:
	statik -src=public