# Define the docker-compose command
DC := cd dev_memorianexus && docker-compose

# Services names
MYSQL_SERVICE := mysql
REDIS_SERVICE := redis
MIGRATION_SERVICE := migration

# Paths
MIGRATE_UP_PATH := migration/migrate_up.sql
MIGRATE_DOWN_PATH := migration/migrate_down.sql

.PHONY: default help test gen-docs bundle\
 		compose-up compose-down compose-re compose-logs\
 		db-migrate-up db-migrate-down db-migrate-re\
 		dev dev-logs dev-shell\
 		build-dev build-app build-staging\
 		cz

# Default to help
default: help

# Show help
help:
	@echo "Available commands:"
	@echo "  make test : Run Go tests"
	@echo "  make compose-up : Start all services with docker-compose"
	@echo "  make compose-down : Stop all services and remove containers"
	@echo "  make compose-logs : Fetch logs for all services"
	@echo "  make db-migrate-up : Perform database migrations"
	@echo "  make db-migrate-down : Rollback database migrations"
	@echo "  make compose-reset : Stop all services and remove data"

test:
	go test ./...

gen-docs:
	# Gen API Documentation
	swag init -g cmd/main.go -o doc/
	# Gen file tree
	@./script/generate_tree_md.sh "bundle,dev_*" > ./doc/PROJECT_STRUCTURE.md

# bundle, @see github.com/bagaking/file_bundle
bundle: gen-docs
	$(MAKE) -f bundle/Makefile
	#file_bundle -v -i ./bundle/_.file_bundle_rc -o ./bundle/_.bundle.txt

# Start services
compose-up:
	$(DC) up -d

# Stop services
compose-down:
	$(DC) down

compose-re: compose-down compose-up

# Display logs
compose-logs:
	$(DC) logs

# Existing makefile content
define run-migration
    $(DC) exec $(MYSQL_SERVICE) bash /migrate.sh $(1)
endef
#$(DC) run -T --rm migrator
#$(DC) run --rm -e MIGRATE_DIRECTION=down migrator

db-migrate-up:
	$(call run-migration,up)

db-migrate-down:
	$(call run-migration,down)

db-migrate-re: db-migrate-down db-migrate-up

# Clean up environment including volumes
compose-reset:
	$(DC) down --volumes

# Build the docker image for our go application
build-app: bundle
	chmod +x ./script/copy_local_modules.sh && sh ./script/copy_local_modules.sh
	$(DC) build app

dev:
	$(DC) up -d app
	@echo Starting app service...
	$(DC) exec -d app /app/memorianexus
	@echo App service has been started in the background.

# Start the built go application
build-dev: build-app compose-re dev

# build, start, follow the log
watch: build-dev dev-logs

# Connect to the app container and follow the logs
dev-logs:
	tail -f ./dev_memorianexus/logs/memorianexus.log -n 1000
	#$(DC) exec app tail -f /app/memorianexus.log

dev-shell:
	$(DC) exec app /bin/sh

build-staging:
	# 编译并上传
	source ./set_env.sh && ./script/build_staging.sh

cz:
	source ./set_env.sh && git cz
