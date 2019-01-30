up:
	docker-compose up -d

down:
	docker-compose down

psql: up
	PGPASSWORD=toto psql -h localhost -U drugstore drugstore

test:
	go test -v github.com/factorysh/drugstore/store