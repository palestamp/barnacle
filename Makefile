




drop-db:
	psql -h localhost -p 5434 -U postgres postgres -c "DROP DATABASE barnacle"
	psql -h localhost -p 5434 -U postgres postgres -c "CREATE DATABASE barnacle"
	migrate -path ./scripts/db/migrations/ -database postgresql://postgres@localhost:5434/barnacle?sslmode=disable up

build:
	go build -o barnacle cmd/barnacle/*.go