# Employees API

REST API сервис для управления данными сотрудников с PostgreSQL, валидацией, логированием и контейнеризацией.

## Возможности

- REST API с версионированием (/v1)
- Валидация данных (ФИО, телефон E.164, город)
- Структурированные JSON логи с request_id и db_time_ms
- PostgreSQL с pgxpool и миграциями
- Docker контейнеризация (45.9MB образ)
- Unit и интеграционные тесты
- Newman/Postman коллекция для API тестирования
- Graceful shutdown

## Требования

- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)
- Newman (опционально, для API тестов)
- make (опционально)

## Быстрый старт

### Способ 1: Автоматический запуск и настройка (рекомендуется)

```bash
make start
```

Эта команда автоматически:
- Проверяет все зависимости
- Останавливает старые контейнеры
- Собирает приложение
- Запускает PostgreSQL и API
- Проверяет healthcheck
- Создает тестового сотрудника

### Способ 2: Ручной запуск через docker-compose

```bash
make run
# или
docker-compose up --build
```

API будет доступен на `http://localhost:8080`

## Доступные команды

### Запуск и управление

```bash
make start            # Быстрый запуск и настройка всей системы
make run              # Запустить через docker-compose
make docker-up        # Поднять сервисы в фоне
make docker-down      # Остановить все сервисы
make build            # Собрать бинарник
make docker-build     # Собрать Docker образ
make clean            # Очистить артефакты
```

### Тестирование

```bash
make test             # Запустить все Go тесты
make test-unit        # Только unit тесты
make test-integration # Только интеграционные тесты
make test-full        # Полное тестирование (unit + integration + API)
make test-newman      # Newman API тесты с детальными логами
```

## API Endpoints

### POST /v1/employees

Создать сотрудника

```bash
curl -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "fullName": "Иван Иванов",
    "phone": "+79991234567",
    "city": "Москва"
  }'
```

Ответ 201:
```json
{
  "id": "c91dd64b-773e-406b-873f-37cc13fa56d5",
  "fullName": "Иван Иванов",
  "phone": "+79991234567",
  "city": "Москва",
  "createdAt": "2025-11-12T08:46:01.794726Z",
  "updatedAt": "2025-11-12T08:46:01.794726Z"
}
```

### GET /v1/employees/{id}

Получить сотрудника по ID

```bash
curl http://localhost:8080/v1/employees/{uuid}
```

Ответ 200:
```json
{
  "id": "c91dd64b-773e-406b-873f-37cc13fa56d5",
  "fullName": "Иван Иванов",
  "phone": "+79991234567",
  "city": "Москва",
  "createdAt": "2025-11-12T08:46:01.794726Z",
  "updatedAt": "2025-11-12T08:46:01.794726Z"
}
```

### GET /v1/healthz

Проверка здоровья сервиса

```bash
curl http://localhost:8080/v1/healthz
```

Ответ 200:
```json
{
  "status": "ok"
}
```

## Валидация

- **fullName**: 2-200 символов, только буквы (кириллица/латиница), пробелы и дефисы
- **phone**: формат E.164 (+[1-15 цифр]), не может начинаться с +0
- **city**: 2-120 символов
- Автоматический тримминг пробелов

## Коды ошибок

- `400` - невалидный JSON, Content-Type или UUID
- `404` - сотрудник не найден
- `405` - метод не поддерживается
- `409` - телефон уже существует
- `422` - ошибка валидации полей
- `500` - внутренняя ошибка сервера

Формат ошибки:
```json
{
  "code": "validation_error",
  "message": "Ошибка валидации",
  "details": {
    "phone": "формат E.164 (+[1-15 цифр])"
  }
}
```

## Newman/Postman тестирование

### Запуск Newman тестов

```bash
make test-newman
# или
./test_newman.sh
```

Эта команда:
- Проверяет доступность API
- Запускает 13 тестов из коллекции
- Показывает детальные логи запросов и ответов
- Сохраняет результаты в `newman_logs/`
- Показывает последние логи с `db_time_ms`

### Импорт в Postman

1. Откройте Postman
2. File → Import
3. Выберите `postman_collection.json`
4. Готово! Коллекция содержит 13 тестов

### Структура тестов

1. **Healthcheck** - проверка работы сервиса (200)
2. **Создание сотрудника** - успешное создание (201)
3. **Получение по ID** - получение существующего сотрудника (200)
4. **Дублирующийся телефон** - ошибка конфликта (409)
5. **Валидация короткого имени** - ошибка валидации (422)
6. **Валидация телефона** - неверный формат (422)
7. **Лишние поля в JSON** - strict JSON parsing (400)
8. **Несуществующий ID** - not found (404)
9. **Невалидный UUID** - bad request (400)
10. **Неподдерживаемый метод** - method not allowed (405)
11. **Неверный Content-Type** - bad request (400)
12. **Граничный случай** - минимальное имя 2 символа (201)
13. **Тримминг пробелов** - автоматическое удаление пробелов (201)

## Логи и мониторинг

### Просмотр логов в реальном времени

```bash
docker-compose logs app -f
```

### Структура логов

Все логи в JSON формате с полями:
- `ts` - timestamp в ISO 8601
- `level` - info/error
- `msg` - тип сообщения (http_request, server_starting, etc.)
- `method` - HTTP метод
- `path` - URL path
- `status` - HTTP статус код
- `latency_ms` - время обработки запроса (мс)
- `db_time_ms` - время выполнения БД запроса (мс) - только при обращении к БД
- `request_id` - ID запроса (X-Request-ID header)
- `remote_addr` - адрес клиента

Пример лога:
```json
{
  "db_time_ms": 5,
  "latency_ms": 20,
  "level": "info",
  "method": "POST",
  "msg": "http_request",
  "path": "/v1/employees",
  "remote_addr": "192.168.65.1:33190",
  "request_id": "",
  "status": 201,
  "ts": "2025-11-12T08:18:26.939110547Z"
}
```

## Детальные отчеты

- `ФИНАЛЬНЫЙ_ОТЧЕТ.md` - полный отчет о проверке системы
- `ЧЕСТНАЯ_ПРОВЕРКА.md` - чек-лист всех требований
- `АНАЛИЗ_ЛОГОВ.md` - детальный анализ логов с примерами
- `newman_logs/` - логи Newman тестов (создаются после `make test-newman`)

## Переменные окружения

- `PORT` - порт сервера (по умолчанию: 8080)
- `POSTGRES_DSN` - строка подключения к PostgreSQL (обязательно)
- `DB_MAX_CONNS` - максимум соединений (по умолчанию: 20)
- `DB_MIN_CONNS` - минимум соединений (по умолчанию: 5)
- `RUN_MIGRATIONS` - запускать ли миграции при старте (по умолчанию: true)

## Структура проекта

```
.
├── cmd/api/              # точка входа
├── internal/
│   ├── config/           # конфигурация
│   ├── database/         # пул и миграции
│   ├── domain/           # модели данных
│   ├── repository/       # работа с БД
│   ├── service/          # бизнес-логика и валидация
│   └── transport/        # HTTP handlers и middleware
├── migrations/           # SQL миграции
├── test/                 # интеграционные тесты
├── postman_collection.json  # Postman коллекция
├── test_newman.sh        # скрипт Newman тестов
├── test_full.sh          # скрипт полного тестирования
├── start_and_setup.sh    # скрипт быстрого запуска
├── Dockerfile            # multi-stage build
├── docker-compose.yml    # оркестрация
└── Makefile             # команды сборки
```

## Технологический стек

- **Go 1.21+** - язык программирования
- **PostgreSQL 14+** - база данных
- **pgx/pgxpool** - драйвер PostgreSQL
- **Docker** - контейнеризация
- **Newman** - автоматизированное API тестирование
- **Testcontainers** - интеграционные тесты

## Архитектура

Чистая архитектура с разделением слоев:
- `cmd/api` - точка входа
- `internal/transport` - HTTP handlers и middleware
- `internal/service` - бизнес-логика и валидация
- `internal/repository` - доступ к БД
- `internal/config` - конфигурация
- `internal/database` - пул соединений и миграции
- `migrations/` - SQL миграции

## Особенности реализации

- Структурированные JSON логи с `db_time_ms` и `latency_ms`
- Автоматические SQL миграции при старте
- DisallowUnknownFields для strict JSON parsing
- Prepared statements для повышения производительности
- Connection pooling с настройкой (min: 5, max: 20)
- Параметризованные запросы (защита от SQL injection)
- Уникальный индекс на телефон
- GIN trgm индекс для полнотекстового поиска
- Тримминг пробелов в полях
- Валидация на уровне сервиса
- Унифицированный формат ошибок
- Минимальный Docker образ (45.9MB)
- PII не логируется (защита персональных данных)

## Разработка

Запуск локально (требует PostgreSQL):

```bash
export POSTGRES_DSN="postgres://user:pass@localhost:5432/employees?sslmode=disable"
go run ./cmd/api
```

Запуск unit тестов:

```bash
make test-unit
```

Запуск интеграционных тестов:

```bash
make test-integration
```

Сборка:

```bash
make build
```

## Полное тестирование

Для полного тестирования всей системы:

```bash
make test-full
```

Эта команда запустит:
1. Unit тесты (17 тестов)
2. Интеграционные тесты (6 тестов)
3. API тесты через curl (7 тестов)
4. Проверку структурированных логов
5. Проверку db_time_ms в логах

## Производительность

Средняя производительность из Newman тестов:
- Среднее время ответа: 13ms
- Минимальное время: 4ms
- Максимальное время: 40ms
- Среднее время DNS lookup: 652µs
- Среднее время первого байта: 10ms

## Лицензия

MIT
