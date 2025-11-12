#!/bin/bash

set -e

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_error() {
    echo -e "${RED}[ОШИБКА]${NC} $1"
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[ШАГ]${NC} $1"
}

echo "==================================="
echo "БЫСТРЫЙ ЗАПУСК И НАСТРОЙКА"
echo "==================================="
echo ""

# Шаг 1: Проверка зависимостей
log_step "1/6 Проверка зависимостей"
MISSING_DEPS=0

if ! command -v docker &> /dev/null; then
    log_error "Docker не установлен"
    MISSING_DEPS=1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose не установлен"
    MISSING_DEPS=1
fi

if ! command -v go &> /dev/null; then
    log_error "Go не установлен"
    MISSING_DEPS=1
fi

if [ $MISSING_DEPS -eq 1 ]; then
    exit 1
fi

log_success "Все зависимости установлены"
echo ""

# Шаг 2: Остановка старых контейнеров
log_step "2/6 Остановка старых контейнеров"
docker-compose down -v >/dev/null 2>&1 || true
log_success "Старые контейнеры остановлены"
echo ""

# Шаг 3: Сборка приложения
log_step "3/6 Сборка приложения"
CGO_ENABLED=0 go build -o api ./cmd/api
log_success "Приложение собрано"
echo ""

# Шаг 4: Запуск docker-compose
log_step "4/6 Запуск docker-compose"
log_info "Запуск PostgreSQL и API сервиса..."
docker-compose up -d

# Ожидание готовности сервисов
log_info "Ожидание готовности сервисов..."
for i in {1..30}; do
    if curl -sf http://localhost:8080/v1/healthz >/dev/null 2>&1; then
        log_success "Сервисы готовы"
        break
    fi

    if [ $i -eq 30 ]; then
        log_error "Таймаут ожидания готовности сервисов"
        docker-compose logs app
        docker-compose down -v
        exit 1
    fi

    sleep 2
done
echo ""

# Шаг 5: Проверка healthcheck
log_step "5/6 Проверка healthcheck"
HEALTH_RESPONSE=$(curl -s http://localhost:8080/v1/healthz)
if echo "$HEALTH_RESPONSE" | grep -q '"status":"ok"'; then
    log_success "Healthcheck: $HEALTH_RESPONSE"
else
    log_error "Healthcheck провалился: $HEALTH_RESPONSE"
    docker-compose down -v
    exit 1
fi
echo ""

# Шаг 6: Создание тестового сотрудника
log_step "6/6 Создание тестового сотрудника"
TEST_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/employees \
    -H "Content-Type: application/json" \
    -d '{"fullName":"Тестовый Пользователь","phone":"+79999999999","city":"Москва"}')

if echo "$TEST_RESPONSE" | grep -q '"id"'; then
    EMPLOYEE_ID=$(echo "$TEST_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    log_success "Тестовый сотрудник создан: ID=$EMPLOYEE_ID"
else
    log_error "Не удалось создать тестового сотрудника"
    echo "Ответ: $TEST_RESPONSE"
fi
echo ""

# Вывод статуса контейнеров
log_info "Статус контейнеров:"
docker-compose ps
echo ""

# Показываем последние логи
log_info "Последние логи приложения:"
docker-compose logs app | tail -10
echo ""

# Итоговая информация
echo "==================================="
echo -e "${GREEN}СИСТЕМА ЗАПУЩЕНА И НАСТРОЕНА${NC}"
echo "==================================="
echo ""
echo "Доступные endpoint'ы:"
echo "  - http://localhost:8080/v1/healthz"
echo "  - http://localhost:8080/v1/employees"
echo "  - http://localhost:8080/v1/employees/{id}"
echo ""
echo "Команды для управления:"
echo "  - docker-compose logs app -f  # Смотреть логи в реальном времени"
echo "  - docker-compose down -v      # Остановить все сервисы"
echo "  - ./test_newman.sh           # Запустить Newman тесты"
echo "  - make test-full             # Полное тестирование"
echo ""
echo "Postman коллекция: postman_collection.json"
echo ""
