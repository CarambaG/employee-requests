# Employee Requests

Веб-приложение для учёта заявок сотрудников.

## Состояние проекта

Первый коммит содержит только инфраструктурную заготовку:

- точку входа HTTP API;
- загрузку конфигурации из переменных окружения;
- корректное завершение приложения;
- endpoint `GET /healthz`;
- Dockerfile;
- PostgreSQL в Docker Compose;
- каталог для SQL-миграций;
- базовый тест HTTP-маршрутизатора.

Бизнес-сущности, база данных и REST API заявок будут добавляться отдельными коммитами.

## Требования

- Go 1.23 или новее;
- Docker и Docker Compose — для запуска PostgreSQL и приложения в контейнерах.

## Локальный запуск

```bash
cp .env.example .env
go run ./cmd/api
```

Проверка:

```bash
curl http://localhost:8080/healthz
```

Ответ:

```json
{"status":"ok","service":"employee-requests"}
```

## Запуск в Docker

```bash
cp .env.example .env
docker compose up --build
```

API будет доступно по адресу `http://localhost:8080`.

## Проверки

```bash
make check
```

## Структура

```text
cmd/api/             точка входа приложения
internal/app/        запуск и остановка приложения
internal/config/     конфигурация
internal/httpapi/    HTTP-маршрутизация и обработчики
migrations/         SQL-миграции
```

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
|---|---:|---|
| `APP_ENV` | `local` | окружение приложения |
| `HTTP_ADDR` | `:8080` | адрес HTTP-сервера |
| `SHUTDOWN_TIMEOUT` | `10s` | таймаут корректного завершения |
| `DATABASE_URL` | — | строка подключения к PostgreSQL |
