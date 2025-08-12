# L0 — Order service
Сервис для приёма заказов из Kafka, сохранения в PostgreSQL (через `ent`) и кэширования в памяти.
В проекте есть HTTP-эндпоинт для получения заказа по `order_id`.

---

## Структура проекта

- `cmd/app` — точка входа приложения
- `config/local.yaml` — параметры локального окружения (env)
- `ent/schema` — схемы ent (хранятся в git)
- `ent/` — сгенерированный ent код (игнорируется в git)
- `internal/storage/postgres` — работа с БД и миграции
- `internal/kafka` — работа с Kafka
- `internal/handlers` — Хэндлеры (контроллеры)
- `githooks/` — git hooks (pre-push)
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

## Тесты

Запускать при работающих контейнерах!
```bash
docker exec -e CONFIG_PATH=/app/config/local.yaml l0-app-1 go test ./...
```
или лучше
```bash
make runTests
```