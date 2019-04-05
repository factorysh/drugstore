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