


# Subscription Management API

Тестовое задание: REST API для управления подписками пользователей. Реализовано на чистом Go без использования фреймворков.

## Особенности реализации

- **Чистый net/http** - без использования веб-фреймворков (Gin, Echo и т.д.)
- **Нативный SQL** - работа с БД через стандартный database/sql
- **Структурированное логирование** - с использованием zap
- **Валидация** - через go-playground/validator
- **Миграции** - с помощью SQL
- **Документация** - автоматическая генерация Swagger

## Технологический стек

- Go 1.23+
- PostgreSQL 15+
- Swagger UI
- Docker

## Запуск проекта

1. Убедитесь, что у вас установлены Docker и Docker Compose
2. Создайте файл `.env` в корне проекта:

```env
DB_HOST=postgres
DB_NAME=Subscription
DB_PORT=5432
DB_PASSWORD=1234
DB_USER=postgres
```

3. Выполните команды:

```bash
# Собрать и запустить контейнеры
docker-compose up --build

# После запуска:
# API будет доступно на http://localhost:8080
# Swagger UI на http://localhost:8080/swagger
```

## API Endpoints

### 1. Создание подписки
**POST** `/api/v1/subscriptions/create-subscription`

Пример запроса:
```json
{
  "user_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
  "service_name": "Netflix",
  "price": 500,
  "start_date": "01-2024",
  "end_date": "12-2024"
}
```

Пример ответа (201 Created):
```json
{
  "id": "b5c6d7e8-f9g0-1234-h5i6-j7k8l9m0n1o2",
  "status": "created"
}
```

### 2. Получение подписок пользователя
**GET** `/api/v1/subscriptions/get-subscription?user-id={user_id}`

Пример ответа (200 OK):
```json
[
  {
    "id": "b5c6d7e8-f9g0-1234-h5i6-j7k8l9m0n1o2",
    "service_name": "Netflix",
    "price": 500,
    "user_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-01T00:00:00Z",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### 3. Обновление подписки
**PUT** `/api/v1/subscriptions/update-subscription`

Пример запроса:
```json
{
  "subscription_id": "b5c6d7e8-f9g0-1234-h5i6-j7k8l9m0n1o2",
  "service_name": "Netflix Premium",
  "price": 700,
  "start_date": "02-2024",
  "end_date": "12-2024"
}
```

Пример ответа (202 Accepted):
```json
{
  "status": "updated"
}
```

### 4. Удаление подписки
**DELETE** `/api/v1/subscriptions/delete-subscription?subscription-id={subscription_id}`

Пример ответа (200 OK):
```json
{
  "status": "deleted"
}
```

### 5. Расчет стоимости подписок
**GET** `/api/v1/subscriptions/calculate-cost?start-date=01-2024&end-date=12-2024&user-id={user_id}`

Пример ответа (200 OK):
```json
{
  "total_cost": 1200,
  "period": {
    "start": "01-2024",
    "end": "12-2024"
  }
}
```

## Структура проекта

```
.
├── cmd/app/main.go          # Точка входа
├── internal
│   ├── handler              # HTTP обработчики
│   ├── service              # Бизнес-логика
│   ├── repository           # Работа с БД
│   ├── models               # Модели данных
│   ├── utils                # Вспомогательные функции
│   └── core                 # Конфигурация и утилиты
├── migrations               # SQL миграции
├── docs                     # Swagger документация
├── docker-compose.yml       # Конфигурация Docker
├── .env                     # Переменные окружения
└── Dockerfile               # Сборка приложения
```

## Работа с базой данных

Подключение к PostgreSQL через стандартный `database/sql`:

```go
func CreateDBConnection() (*sql.DB, error) {
    cfg := configs.Init()
    db, err := sql.Open("postgres", cfg.DB.DBUrl())
    if err != nil {
        return nil, err
    }
    return db, db.Ping()
}
```

Пример SQL запроса (репозиторий):
```go
query = `SELECT id, service_name, price FROM subscriptions WHERE user_id = $1`
rows, err := r.db.QueryContext(ctx, query, userID)
```

## Дополнительная информация

- Все ошибки логируются с полным контекстом
- Подробная документация доступна через Swagger UI
- Для тестирования можно использовать Swagger UI по адресу http://localhost:8080/swagger

