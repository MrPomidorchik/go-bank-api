# Go Bank API

Банковский сервис — REST API на Go. Пользователи могут создавать банковские счета, карты и управлять ими, проводить транзакции, а также оформлять кредит. Доступна информация о транзакциях, график платежей по кредиту и аналитика финансов. Оповещения о транзакциях приходят на почту, а ключевая ставка, для оформления кредита, формируется с учетом ключевой ставки ЦБ РФ. 

## Стек

- Go 1.26
- PostgreSQL 18
- gorilla/mux 1.8.1 — маршрутизация
- lib/pq 1.12.3 — работа с БД
- x/crypto 0.51.0 — шифрование
- golang-jwt/jwt/v5 5.3.1 — JWT-аутентификация
- go-mail/mail/v2 2.3.0 — работа с SMTP
- sirupsen/logrus 1.9.4 — логирование
- beevik/etree 1.6.0 — парсинг XML (интеграция API ЦБ РФ)
- Docker

## Быстрый старт

### Через Docker Compose, который поднимает и базу, и приложение:

Создайте .env.docker (см. .env.example, DB_HOST=db)

```bash
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`

```bash
# Остановить проект
docker compose down
```

```bash
# Полное удаление контейнеров и данных PostgreSQL
docker compose down -v
```

### Локальный запуск

Создайте .env (см. .env.example, DB_HOST=localhost)

```bash
# 1. Установка зависимостей
go mod tidy
```

```bash
# 2. Запуск миграций
Get-ChildItem .\migrations\*.sql | Sort-Object Name | ForEach-Object {
psql -U postgres -d bank_api -f $_.FullName
}
```

```bash
# 3. Запуск приложения
go run ./cmd/api
```

Сервис будет доступен на `http://localhost:8080`

## API

### Аутентификация

| Метод | URL | Описание |
|-------|-----|----------|
| POST | `/register` | Регистрация нового пользователя |
| POST | `/login` | Вход, возвращает JWT-токен |

### Счета (требуют `Authorization: Bearer <token>`)

| Метод | URL | Описание |
|-------|-----|----------|
| POST  | `/accounts` | Создать банковский счёт |
| GET   | `/accounts` | Получить список своих счетов |
| POST  | `/accounts/{accountId}/deposit` | Пополнить счёт |
| POST  | `/accounts/{accountId}/withdraw` | Списать деньги со счёта |
| POST  | `/transfer` | Перевод между счетами |

### Банковские карты (требуют `Authorization: Bearer <token>`)

| Метод | URL                      | Описание                   |
|-------|--------------------------|----------------------------|
| POST  | `/cards`                 | Выпустить карту            |
| GET  | `/cards` | Получить список своих карт |
| POST  | `/cards/pay` | Оплата картой            |

### Банковские кредиты (требуют `Authorization: Bearer <token>`)

| Метод | URL | Описание |
|-------|-----|----------|
| POST | `/credits` | Оформить кредит |
| GET | `/credits` | Получить список своих кредитов |
| GET | `/credits/{creditId}/schedule` | Получить график платежей по кредиту |

### Финансовая аналитика (требуют `Authorization: Bearer <token>`)

| Метод | URL | Описание |
|-------|-----|----------|
| GET   | `/analytics` | Финансовая аналитика за текущий месяц |
| GET   | `/accounts/{accountId}/predict?days=30` | Прогноз баланса счёта |

### История операций (требуют `Authorization: Bearer <token>`)

| Метод | URL                             | Описание                                |
|-------|---------------------------------|-----------------------------------------|
| GET   | `/transactions`                 | Получить все операции пользователя      |
| GET   | `/transactions?account_id=uuid` | Операции по конкретному счёту           |
| GET   | `/transactions?type=operation`  | Операции по типу (deposit или withdraw) |
| GET   | `/transactions?date_from=2026-05-01&date_to=2026-05-31` | Операции за период           |
| GET   | `/transactions?account_id=uuid&type=card_payment&date_from=2026-05-01&date_to=2026-05-31` | Комбинированный фильтр           |

### Утилитарные (требуют `Authorization: Bearer <token>`)

| Метод | URL                      | Описание                              |
|-------|--------------------------|---------------------------------------|
| GET | `/me`| Проверить текущего пользователя по JWT |
| POST | `/notifications/test-email` | Отправить тестовое оповещение на почту|
| POST | `/credits/process-payments` | Запустить обработку просроченных/плановых платежей |
| GET | `/rates/cbr` | Получить ключевую ставку ЦБ РФ + банковскую маржу |

## Примеры запросов

```bash
# Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{ "email": "test@mail.ru", "username": "testuser", "password": "123456" }'

# Вход
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{ "email": "test@mail.ru", "password": "123456" }'

# Проверка текущего пользователя
curl -X GET http://localhost:8080/me \
  -H "Authorization: Bearer TOKEN"

# Создание нового счета (ограничение: только в рублях)
curl -X POST http://localhost:8080/accounts \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "currency": "RUB" }'

# Получить счета пользователя
curl -X GET http://localhost:8080/accounts \
  -H "Authorization: Bearer TOKEN"
  
# Пополнить счет
curl -X POST http://localhost:8080/accounts/ACCOUNT_ID/deposit \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "amount": 1000 }'
  
# Списать деньги со счета
curl -X POST http://localhost:8080/accounts/ACCOUNT_ID/withdraw \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "amount": 300 }'
  
# Перевод между счетами
curl -X POST http://localhost:8080/transfer \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "from_account_id": "ACCOUNT_ID", "to_account_number": "ACCOUNT_NUMBER", "amount": 500 }'
  
# Выпустить карту (после выпуска карты, сохраните полученный в ответе cvv-код)
curl -X POST http://localhost:8080/cards \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "account_id": "ACCOUNT_ID" }'
  
# Получить карты пользователя
curl -X GET http://localhost:8080/cards \
  -H "Authorization: Bearer TOKEN"

# Оплата картой
curl -X POST http://localhost:8080/cards/pay \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "card_id": "CARD_ID", "cvv": "123", "amount": 500 }'

# Оформить кредит (если необходимо оформить со ставкой ЦБ РФ + маржа — удалить "annual_rate"
curl -X POST http://localhost:8080/credits \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "account_id": "ACCOUNT_ID", "amount": 100000, "term_months": 12, "annual_rate": 20 }'

# Получить список кредитов пользователя
curl -X GET http://localhost:8080/credits \
  -H "Authorization: Bearer TOKEN"

# Получить график платежей по кредиту
curl -X GET http://localhost:8080/credits/CREDIT_ID/schedule \
  -H "Authorization: Bearer TOKEN"
  
# Ручной запуск обработки кредитных платежей
curl -X POST http://localhost:8080/credits/process-payments \
  -H "Authorization: Bearer TOKEN"
  
# Получить финансовую аналитику
curl -X GET http://localhost:8080/analytics \
  -H "Authorization: Bearer TOKEN"
  
# Прогноз баланса на 30 дней
curl -X GET "http://localhost:8080/accounts/ACCOUNT_ID/predict?days=30" \
  -H "Authorization: Bearer TOKEN"
  
# История всех операций (используйте фильтры при необходимости)
curl -X GET http://localhost:8080/transactions \
  -H "Authorization: Bearer TOKEN"
  
# Получить ставку ЦБ РФ + маржу банка
curl -X GET http://localhost:8080/rates/cbr \
  -H "Authorization: Bearer TOKEN"
  
# Отправить тестовое email-уведомление
curl -X POST http://localhost:8080/notifications/test-email \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "to": "your_email@gmail.com" }'
```

## Архитектура

```
bank-api/
├── cmd/
│   └── api/
│       └── main.go     — точка входа
│
├── internal/
│   ├── config/         — конфигурация
│   ├── models/         — моделеи данных
│   ├── repository/     — работа с базой данных
│   ├── service/        — бизнес-логика
│   ├── handler/        — HTTP-запросы
│   ├── middleware/     — промежуточные обработчики HTTP-запросов
│   ├── integration/ 
│   │   ├── cbr/        — интеграция с ЦБ РФ
│   │   └── smtp/       — интеграция с SMTP
│   ├── crypto/         — криптографические утилиты
│   ├── scheduler/      — фоновые задачи
│   ├── validator/      — валидация данных
│   └── apperror/       — единые ошибоки приложения
│
├── migrations/         — SQL-миграции
```
