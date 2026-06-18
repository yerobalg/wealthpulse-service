-include .env

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

migrate:
	goose -dir docs/sql postgres "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USERNAME) password=$(DB_PASSWORD) dbname=$(DB_DBNAME) sslmode=disable" up

swagger:
	swag init -o ./docs/api

# Generate a bcrypt hash for a password, e.g.:
#   make gen-password password=admin123 cost=8
# cost defaults to PASSWORD_SALT_ROUND (or 10 if unset).
gen-password:
	go run ./scripts/genpassword -password=$(password) $(if $(cost),-cost=$(cost),)

# Rename the Go module path across the project, e.g.:
#   make rename-module module=github.com/your-org/your-repo
rename-module:
	./scripts/rename_module.sh $(module)

run:
	go run main.go

build:
	go build -o bin/go-api-boilerplate main.go
