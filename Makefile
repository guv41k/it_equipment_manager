include .env
export

export PROJECT_ROOT=${shell pwd}

.PHONY: env-up env-down env-ps env-logs env-cleanup env-port-forward env-port-close migrate-create migrate-up migrate-down migrate-action

env-up:
	@docker compose up -d managerapp-postgres managerapp-kafka managerapp-clickhouse

env-down:
	@docker compose down

env-logs:
	@docker compose logs -f --tail=100

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose rm -sf port-forwarder

env-cleanup:
	@printf 'Удалить данные PostgreSQL, Kafka и ClickHouse из out/? [y/N]: '; \
	read answer; \
	if [ "$$answer" = "y" ]; then \
		docker compose down && \
		rm -rf out/pgdata out/kafka out/clickhouse && \
		printf 'Данные удалены\n'; \
	fi

migrate-create:
	@if [ -z "$(seq)" ]; then \
		printf 'Укажите имя миграции: make migrate-create seq=init\n'; \
		exit 1; \
	fi
	@mkdir -p migrations
	@docker compose run --rm managerapp-postgres-migrate \
		create -ext sql -dir /migrations -seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down steps=1

migrate-action:
	@if [ -z "$(action)" ]; then \
		printf 'Укажите действие: make migrate-action action=up\n'; \
		exit 1; \
	fi
	@docker compose up --wait managerapp-postgres
	@docker compose run --rm managerapp-postgres-migrate \
		-path /migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@managerapp-postgres:5432/$(POSTGRES_DB)?sslmode=disable" \
		"$(action)" $(steps)
