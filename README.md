# Employee Requests

Веб-приложение для учёта заявок сотрудников.

## Реализовано

- HTTP-сервер и endpoint `GET /health`;
- проверка доступности PostgreSQL через health-check;
- конфигурация приложения и пула соединений через переменные окружения;
- подключение к PostgreSQL через `pgxpool`;
- корректное завершение HTTP-сервера и пула соединений;
- объектная модель сотрудников и заявок;
- бизнес-правила переходов статусов;
- нормализованная схема PostgreSQL;
- PostgreSQL-репозиторий сотрудников с операциями создания, чтения, изменения и удаления;
- REST API справочника сотрудников с валидацией и единым форматом ошибок;
- последовательное применение SQL-миграций;
- Dockerfile и Docker Compose.

## Бизнес-процесс

Допустимы только последовательные переходы:

```text
Новая -> В работе -> Выполнена
```

Переход из статуса `Новая` сразу в `Выполнена`, возврат на предыдущий статус и изменение статуса выполненной заявки запрещены. Правило реализовано методом `Request.ChangeStatus` и покрыто модульными тестами.

## Структура базы данных

- `departments` — подразделения;
- `positions` — должности;
- `employees` — сотрудники со ссылками на подразделение и должность;
- `request_statuses` — справочник статусов;
- `request_status_transitions` — допустимые переходы между статусами;
- `requests` — заявки со ссылками на автора, исполнителя и статус;
- `schema_migrations` — применённые миграции.

Подразделения, должности и статусы вынесены в отдельные таблицы, поэтому повторяющиеся значения не хранятся в каждой записи сотрудника или заявки. Целостность связей обеспечивают внешние ключи. Номер заявки используется как первичный ключ и генерируется PostgreSQL.

Составной индекс для поиска просроченных заявок пока намеренно не добавлен. Сначала необходимо загрузить тестовые данные и зафиксировать исходный план и время запроса, а затем добавить индекс отдельным оптимизационным коммитом.

## REST API сотрудников

| Метод | Путь | Назначение |
|---|---|---|
| `POST` | `/api/v1/employees` | создать сотрудника |
| `GET` | `/api/v1/employees` | получить список сотрудников |
| `GET` | `/api/v1/employees/{id}` | получить сотрудника |
| `PUT` | `/api/v1/employees/{id}` | полностью изменить сотрудника |
| `DELETE` | `/api/v1/employees/{id}` | удалить сотрудника |

Для создания сотрудника в таблицах `departments` и `positions` должны существовать соответствующие записи. До добавления отдельного API справочников тестовые записи можно создать напрямую:

```bash
docker compose exec db psql \
  -U employee_requests \
  -d employee_requests \
  -c "INSERT INTO departments (name) VALUES ('Разработка'); INSERT INTO positions (name) VALUES ('Инженер-программист');"
```

Создание сотрудника:

```bash
curl -i -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Иванов Иван Иванович",
    "department_id": 1,
    "position_id": 1
  }'
```

Успешный ответ имеет статус `201 Created`, заголовок `Location` и тело:

```json
{
  "id": 1,
  "full_name": "Иванов Иван Иванович",
  "department": {
    "id": 1,
    "name": "Разработка"
  },
  "position": {
    "id": 1,
    "name": "Инженер-программист"
  }
}
```

Получение списка:

```bash
curl http://localhost:8080/api/v1/employees
```

Изменение сотрудника:

```bash
curl -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Петров Пётр Петрович",
    "department_id": 1,
    "position_id": 1
  }'
```

Удаление сотрудника:

```bash
curl -i -X DELETE http://localhost:8080/api/v1/employees/1
```

API использует единый формат ошибок:

```json
{
  "error": {
    "code": "not_found",
    "message": "employee not found"
  }
}
```

Основные HTTP-коды: `400` для некорректных данных, `404` для отсутствующего сотрудника, `409` для конфликта со связанными данными и `500` для внутренних ошибок.

## Сервис и репозиторий сотрудников

`internal/storage/postgres/employee_repository.go` содержит операции:

- `Create`;
- `GetByID`;
- `List`;
- `Update`;
- `Delete`.

Сервис `internal/employee` проверяет входные данные и не зависит от HTTP или PostgreSQL. Репозиторий возвращает сотрудника вместе с данными подразделения и должности. Ошибка отсутствующей записи преобразуется в `domain.ErrNotFound`, а нарушения внешних ключей и уникальности — в `domain.ErrConflict`.

## Требования

- Go 1.25 или новее;
- Docker и Docker Compose.

## Запуск через Docker

```bash
cp .env.example .env
docker compose up --build
```

Docker Compose последовательно:

1. запускает PostgreSQL;
2. применяет ещё не выполненные SQL-миграции;
3. запускает API.

Проверка:

```bash
curl http://localhost:8080/health
```

Ответ при доступной базе данных:

```json
{"status":"ok","service":"employee-requests","database":"available"}
```

## Локальный запуск

Сначала запустите базу данных и миграции:

```bash
docker compose up -d db
docker compose run --rm migrate
```

Загрузите переменные окружения и запустите API:

```bash
set -a
. ./.env
set +a
go run ./cmd/api
```

## Повторное применение миграций

```bash
make migrate-up
```

Уже выполненные миграции пропускаются на основании таблицы `schema_migrations`.

## Проверки

```bash
make check
```

## Структура проекта

```text
cmd/api/                       точка входа HTTP API
internal/app/                  жизненный цикл приложения
internal/config/               конфигурация
internal/domain/               объектная модель и общие ошибки
internal/employee/             бизнес-логика сотрудников
internal/httpapi/              HTTP-маршрутизация и обработчики
internal/storage/postgres/     пул соединений и PostgreSQL-репозитории
migrations/                    SQL-миграции и скрипт их применения
```
