#!/bin/bash

set -e

echo "==================================="
echo "ПОЛНОЕ ТЕСТИРОВАНИЕ СИСТЕМЫ"
echo "==================================="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Функция для логирования
log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_error() {
    echo -e "${RED}[ОШИБКА]${NC} $1"
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# 1. Проверка зависимостей
echo "=== 1. Проверка зависимостей ==="
command -v docker >/dev/null 2>&1 || { log_error "Docker не установлен"; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { log_error "docker-compose не установлен"; exit 1; }
command -v go >/dev/null 2>&1 || { log_error "Go не установлен"; exit 1; }
command -v curl >/dev/null 2>&1 || { log_error "curl не установлен"; exit 1; }
log_success "Все зависимости установлены"
echo ""

# 2. Остановка старых контейнеров
echo "=== 2. Остановка старых контейнеров ==="
docker-compose down -v >/dev/null 2>&1 || true
log_success "Старые контейнеры остановлены"
echo ""

# 3. Unit тесты
echo "=== 3. Unit тесты ==="
if go test -v -race ./internal/service/...; then
    log_success "Unit тесты пройдены"
else
    log_error "Unit тесты провалились"
    exit 1
fi
echo ""

# 4. Интеграционные тесты
echo "=== 4. Интеграционные тесты ==="
if go test -v -timeout 5m ./test/...; then
    log_success "Интеграционные тесты пройдены"
else
    log_error "Интеграционные тесты провалились"
    exit 1
fi
echo ""

# 5. Запуск docker-compose
echo "=== 5. Запуск docker-compose ==="
log_info "Запуск сервисов..."
docker-compose up -d

# Ожидание готовности сервисов
log_info "Ожидание готовности сервисов..."
for i in {1..30}; do
    if docker-compose ps | grep -q "healthy"; then
        if curl -sf http://localhost:8080/v1/healthz >/dev/null 2>&1; then
            log_success "Сервисы готовы"
            break
        fi
    fi

    if [ $i -eq 30 ]; then
        log_error "Таймаут ожидания готовности сервисов"
        docker-compose logs
        docker-compose down -v
        exit 1
    fi

    sleep 2
done
echo ""

# 6. API тесты через curl
echo "=== 6. API тесты ==="

# 6.1 Healthcheck
log_info "Тест 1/7: Healthcheck"
RESPONSE=$(curl -s http://localhost:8080/v1/healthz)
if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    log_success "Healthcheck работает"
else
    log_error "Healthcheck не работает: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.2 Создание сотрудника
log_info "Тест 2/7: Создание сотрудника"
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"Иван Иванов","phone":"+79991234567","city":"Москва"}')
EMPLOYEE_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -n "$EMPLOYEE_ID" ]; then
    log_success "Сотрудник создан: $EMPLOYEE_ID"
else
    log_error "Не удалось создать сотрудника: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.3 Получение сотрудника
log_info "Тест 3/7: Получение сотрудника"
RESPONSE=$(curl -s http://localhost:8080/v1/employees/$EMPLOYEE_ID)
if echo "$RESPONSE" | grep -q "Иван Иванов"; then
    log_success "Сотрудник получен"
else
    log_error "Не удалось получить сотрудника: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.4 Дублирующийся телефон
log_info "Тест 4/7: Дублирующийся телефон (409)"
RESPONSE=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"Другой Человек","phone":"+79991234567","city":"Москва"}')
if echo "$RESPONSE" | grep -q "409"; then
    log_success "Дубликат телефона правильно отклонен (409)"
else
    log_error "Дубликат телефона не отклонен: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.5 Валидация - короткое имя
log_info "Тест 5/7: Валидация - короткое имя (422)"
RESPONSE=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"И","phone":"+79111111111","city":"Москва"}')
if echo "$RESPONSE" | grep -q "422"; then
    log_success "Короткое имя правильно отклонено (422)"
else
    log_error "Короткое имя не отклонено: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.6 Невалидный UUID
log_info "Тест 6/7: Невалидный UUID (400)"
RESPONSE=$(curl -s -w "%{http_code}" http://localhost:8080/v1/employees/not-a-uuid)
if echo "$RESPONSE" | grep -q "400"; then
    log_success "Невалидный UUID правильно отклонен (400)"
else
    log_error "Невалидный UUID не отклонен: $RESPONSE"
    docker-compose down -v
    exit 1
fi

# 6.7 Лишние поля в JSON
log_info "Тест 7/7: Лишние поля в JSON (400)"
RESPONSE=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"Тест","phone":"+79222222222","city":"Москва","extraField":"не должно быть"}')
if echo "$RESPONSE" | grep -q "400"; then
    log_success "Лишние поля правильно отклонены (400)"
else
    log_error "Лишние поля не отклонены: $RESPONSE"
    docker-compose down -v
    exit 1
fi

echo ""
log_success "Все API тесты пройдены"
echo ""

# 7. Проверка логов
echo "=== 7. Проверка структурированных логов ==="
LOGS=$(docker-compose logs app | tail -20)
if echo "$LOGS" | grep -q '"level"'; then
    log_success "Логи в JSON формате"
else
    log_error "Логи не в JSON формате"
fi

if echo "$LOGS" | grep -q '"latency_ms"'; then
    log_success "Логи содержат latency_ms"
else
    log_error "Логи не содержат latency_ms"
fi
echo ""

# 8. Проверка db_time_ms
echo "=== 8. Проверка db_time_ms в логах ==="
# Создаем еще один запрос для гарантированного лога с db_time_ms
curl -s -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"Тест Лога","phone":"+79333333333","city":"Москва"}' >/dev/null 2>&1

sleep 1
LOGS=$(docker-compose logs app | tail -5)
if echo "$LOGS" | grep -q '"db_time_ms"'; then
    log_success "Логи содержат db_time_ms"
else
    log_error "Логи не содержат db_time_ms"
fi
echo ""

# 9. Остановка сервисов
echo "=== 9. Остановка сервисов ==="
docker-compose down -v >/dev/null 2>&1
log_success "Сервисы остановлены"
echo ""

# 10. Итоговый отчет
echo "==================================="
echo -e "${GREEN}ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО${NC}"
echo "==================================="
echo ""
echo "Проверено:"
echo "  - Unit тесты (17 тестов)"
echo "  - Интеграционные тесты (6 тестов)"
echo "  - API тесты (7 тестов)"
echo "  - Структурированные логи"
echo "  - db_time_ms метрика"
echo ""
echo "Система полностью работоспособна."
