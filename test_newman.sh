#!/bin/bash

set -e

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

log_request() {
    echo -e "${BLUE}[ЗАПРОС]${NC} $1"
}

log_response() {
    echo -e "${CYAN}[ОТВЕТ]${NC} $1"
}

echo "==================================="
echo "NEWMAN API ТЕСТЫ С ДЕТАЛЬНЫМИ ЛОГАМИ"
echo "==================================="
echo ""

# Проверка newman
if ! command -v newman &> /dev/null; then
    log_error "Newman не установлен. Установите: npm install -g newman"
    exit 1
fi

# Проверка доступности API
log_info "Проверка доступности API..."
if ! curl -sf http://localhost:8080/v1/healthz >/dev/null 2>&1; then
    log_error "API недоступен на http://localhost:8080"
    log_info "Запустите сервисы: docker-compose up -d"
    exit 1
fi
log_success "API доступен"
echo ""

# Создаем временную директорию для логов
LOG_DIR="./newman_logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$LOG_DIR/newman_${TIMESTAMP}.log"

log_info "Логи будут сохранены в: $LOG_FILE"
echo ""

# Запуск newman с детальными логами
log_info "Запуск Newman тестов..."
echo ""

newman run postman_collection.json \
    --reporters cli,json \
    --reporter-json-export "$LOG_DIR/newman_${TIMESTAMP}.json" \
    --verbose \
    --color on \
    --delay-request 500 2>&1 | tee "$LOG_FILE"

# Проверка результата
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    log_success "Все Newman тесты пройдены успешно"
    echo ""

    # Показываем сводку
    echo "==================================="
    echo "СВОДКА ТЕСТИРОВАНИЯ"
    echo "==================================="

    # Подсчет тестов из лога
    TOTAL_TESTS=$(grep -c "✓" "$LOG_FILE" 2>/dev/null || echo "0")
    FAILED_TESTS=$(grep -c "✖" "$LOG_FILE" 2>/dev/null || echo "0")

    echo "Всего тестов: $TOTAL_TESTS"
    echo "Успешных: $TOTAL_TESTS"
    echo "Провалившихся: $FAILED_TESTS"
    echo ""
    echo "Детальные логи: $LOG_FILE"
    echo "JSON отчет: $LOG_DIR/newman_${TIMESTAMP}.json"
    echo ""

    # Показываем последние логи приложения
    log_info "Последние логи приложения с db_time_ms:"
    echo ""
    docker-compose logs app 2>/dev/null | grep '"db_time_ms"' | tail -5 | while IFS= read -r line; do
        # Извлекаем JSON и форматируем
        json_part=$(echo "$line" | grep -o '{.*}')
        if [ -n "$json_part" ]; then
            echo "$json_part" | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"  {data.get('method','')} {data.get('path','')} -> {data.get('status','')} (latency: {data.get('latency_ms','')}ms, db: {data.get('db_time_ms','')}ms)\")" 2>/dev/null || echo "  $json_part"
        fi
    done
    echo ""

else
    echo ""
    log_error "Некоторые тесты провалились"
    echo ""
    echo "Проверьте логи: $LOG_FILE"
    exit 1
fi
