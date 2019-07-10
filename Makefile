GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)
.PHONY: statik
TEST_OPTION?=-v

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
	go test $(TEST_OPTION) github.com/factorysh/drugstore/conf
	go test $(TEST_OPTION) github.com/factorysh/drugstore/rest
	go test $(TEST_OPTION) github.com/factorysh/drugstore/schema
	go test $(TEST_OPTION) github.com/factorysh/drugstore/store
	go test $(TEST_OPTION) github.com/factorysh/drugstore/rpc

statik:
	statik -src=public

node_modules:
	npm install

public/css/bulma.css: node_modules
	cp node_modules/bulma/css/bulma.css public/css/bulma.css

assets: public/css/bulma.css