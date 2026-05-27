ENV_FILE_PATH := .env
COMPOSE_FILE := docker-compose.yaml
COMPOSE_FLAGS := --env-file $(ENV_FILE_PATH)
PGDATA_DIR := ./_data/pgdata

# =============================================================================
# БАЗА ДАННЫХ
# =============================================================================
.PHONY: db-up db-down db-logs db-restart db-ps

## db-up: Запустить БД
db-up:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) up -d db

## db-down: Остановить и удалить контейнеры БД
db-down:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) down

## db-restart: Перезапустить БД
db-restart:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) restart db

## db-logs: Логи БД
db-logs:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) logs -f db

## db-ps: Статус БД
db-ps:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) ps

## db-shell: зайти в контейнер
db-shell:
	docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) exec db sh


# =============================================================================
# ОЧИСТКА
# =============================================================================
.PHONY: prune clean

prune:
	docker system prune -af --volumes

clean:
	@read -p "⚠️ Удалить ВСЕ данные Postgres? [y/N] " confirm && \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		docker compose -f $(COMPOSE_FILE) $(COMPOSE_FLAGS) down -v && \
		sudo rm -rf $(PGDATA_DIR) && \
		echo "OK cleaned"; \
	else \
		echo "cancelled"; \
	fi

# =============================================================================
# SWAGGER
# =============================================================================
.PHONY: swagger

swagger:
	swag init -g cmd/server/main.go --parseInternal