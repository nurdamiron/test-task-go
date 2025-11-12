#!/bin/bash

echo "=== Граничные случаи валидации ==="
echo ""

echo "1. Максимальная длина ФИО (200 символов):"
NAME200=$(printf 'А%.0s' {1..200})
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d "{\"fullName\":\"$NAME200\",\"phone\":\"+11234567892\",\"city\":\"Город\"}")
echo "$RESPONSE" | jq -c '{created: (.id != null), nameLength: (.fullName // "" | length)}'

echo ""
echo "2. Превышение максимальной длины ФИО (201 символ):"
NAME201=$(printf 'А%.0s' {1..201})
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d "{\"fullName\":\"$NAME201\",\"phone\":\"+11234567893\",\"city\":\"Город\"}" | jq -c

echo ""
echo "3. Минимальная длина города (2 символа):"
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"fullName":"Тест Тестов","phone":"+11234567894","city":"Уф"}' | jq -c '{created: (.id != null)}'

echo ""
echo "4. Максимальная длина города (120 символов):"
CITY120=$(printf 'А%.0s' {1..120})
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d "{\"fullName\":\"Тест Максимум\",\"phone\":\"+11234567895\",\"city\":\"$CITY120\"}" | jq -c '{created: (.id != null)}'

echo ""
echo "5. Превышение максимальной длины города (121 символ):"
CITY121=$(printf 'А%.0s' {1..121})
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d "{\"fullName\":\"Тест Максимум2\",\"phone\":\"+11234567896\",\"city\":\"$CITY121\"}" | jq -c

echo ""
echo "6. Минимальный валидный телефон E.164 (+1X):"
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"fullName":"Короткий Номер","phone":"+12","city":"Москва"}' | jq -c

echo ""
echo "7. Максимальный телефон E.164 (15 цифр после +):"
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"fullName":"Длинный Номер","phone":"+123456789012345","city":"Москва"}' | jq -c '{created: (.id != null)}'

echo ""
echo "8. Превышение максимального телефона (16 цифр):"
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"fullName":"Слишком Длинный","phone":"+1234567890123456","city":"Москва"}' | jq -c

echo ""
echo "9. Пробелы в начале и конце (тримминг):"
curl -s -X POST http://localhost:8080/v1/employees \
  -H "Content-Type: application/json" \
  -d '{"fullName":"  Пробелы Вокруг  ","phone":"+11234567897","city":"  Город  "}' | jq -c '{fullName: .fullName, city: .city}'
