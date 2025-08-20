# L0 — Order service
Сервис для приёма заказов из Kafka, сохранения в PostgreSQL (через `ent`) и кэширования в памяти.
В проекте есть HTTP-эндпоинт для получения заказа по `order_id`.

---

## Структура проекта

- `cmd/app` — точка входа приложения
- `config/local.yaml` — параметры локального окружения (env)
- `internal/repository/postgres` — работа с БД и миграции
- `internal/cache` — работа с кэшом
- `internal/config` — настройки конфига
- `internal/handlers` — хэндлеры (контроллеры)
- `internal/models` — работа с моделями
- `internal/kafka` — работа с Kafka
- `internal/service` — бизнес-логика
- `test/` — тесты
- `githooks/` — git hooks (pre-push, pre-commit)
- `docker-compose.yaml` — локальное окружение (Postgres, Kafka и пр.)
- `Makefile` — команды для разработки и запуска

---

## Требования

- Go 1.23+
- Docker & Docker Compose

---

## Быстрый старт

```bash
git clone <repo>
cd <repo>
make setup
docker compose up -d --build
```

## Миграции

**Запускать при работающих контейнерах!**
```bash
make migrate
```

## Тесты

**Запускать при работающих контейнерах!**
```bash
docker exec -e CONFIG_PATH=/app/config/local.yaml l0-app-1 go test ./...
```
или лучше
```bash
make runTests
```