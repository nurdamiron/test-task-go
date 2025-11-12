.PHONY: run test test-unit test-integration test-full test-newman start clean build docker-build docker-up docker-down help

help:
	@echo "Доступные команды:"
	@echo "  make start            - быстрый запуск и настройка всей системы"
	@echo "  make run              - запустить приложение через docker-compose"
	@echo "  make test             - запустить все тесты"
	@echo "  make test-unit        - запустить unit тесты"
	@echo "  make test-integration - запустить интеграционные тесты"
	@echo "  make test-full        - полное тестирование (unit + integration + API)"
	@echo "  make test-newman      - запустить Newman API тесты с детальными логами"
	@echo "  make build            - собрать бинарник"
	@echo "  make docker-build     - собрать Docker образ"
	@echo "  make docker-up        - поднять сервисы через docker-compose"
	@echo "  make docker-down      - остановить сервисы"
	@echo "  make clean            - очистить артефакты сборки"

start:
	@echo "Быстрый запуск и настройка системы..."
	./start_and_setup.sh

run:
	@echo "Запуск приложения через docker-compose..."
	docker-compose up --build

test:
	@echo "Запуск всех тестов..."
	go test -v -race -timeout 5m ./...

test-unit:
	@echo "Запуск unit тестов..."
	go test -v -race ./internal/service/...

test-integration:
	@echo "Запуск интеграционных тестов..."
	go test -v -timeout 5m ./test/...

test-full:
	@echo "Полное тестирование системы..."
	./test_full.sh

test-newman:
	@echo "Запуск Newman тестов..."
	./test_newman.sh

build:
	@echo "Сборка приложения..."
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

docker-build:
	@echo "Сборка Docker образа..."
	docker build -t employees-api:latest .

docker-up:
	@echo "Запуск сервисов..."
	docker-compose up -d

docker-down:
	@echo "Остановка сервисов..."
	docker-compose down

clean:
	@echo "Очистка артефактов..."
	rm -rf bin/
	docker-compose down -v
