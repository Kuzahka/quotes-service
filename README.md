## 🚀 Быстрый старт

### Вариант 1: Docker Compose (рекомендуется)

```bash
# Клонировать репозиторий
git clone https://github.com/Kuzahka/quotes-service
cd quotes-service

# Запустить все сервисы
docker-compose up -d

# Проверить статус
docker-compose ps

# Просмотреть логи
docker-compose logs -f quotes-service
```

### Вариант 2: Локальная разработка

```bash
# Установить зависимости
go mod tidy

# Запустить PostgreSQL
docker-compose up -d postgres

# Запустить приложение
make run

# Или напрямую
go run cmd/server/main.go
```

## 📡 API Endpoints

### Создание цитаты
```bash
curl -X POST http://localhost:8080/quotes \
  -H "Content-Type: application/json" \
  -d '{
    "author": "Confucius",
    "quote": "Life is simple, but we insist on making it complicated."
  }'
```

**Ответ:**
```json
{
  "data": {
    "id": 1,
    "author": "Confucius",
    "quote": "Life is simple, but we insist on making it complicated.",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Получение всех цитат
```bash
curl http://localhost:8080/quotes
```

### Получение случайной цитаты
```bash
curl http://localhost:8080/quotes/random
```

### Фильтрация по автору
```bash
curl "http://localhost:8080/quotes?author=Confucius"
```

### Удаление цитаты
```bash
curl -X DELETE http://localhost:8080/quotes/1
```

### Health Check
```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "database": "connected"
  }
}
```

## 🏗️ Архитектура

### Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│               HTTP Handler              │  ← Presentation Layer
├─────────────────────────────────────────┤
│              Quote Service              │  ← Business Logic Layer  
├─────────────────────────────────────────┤
│            Repository Interface         │  ← Interface Layer
├─────────────────────────────────────────┤
│          PostgreSQL Repository          │  ← Data Access Layer
└─────────────────────────────────────────┘
```

### Основные компоненты

1. **Domain Layer** (`internal/domain/`)
   - Содержит бизнес-модели и интерфейсы
   - Не зависит от внешних библиотек

2. **Service Layer** (`internal/service/`)
   - Реализует бизнес-логику
   - Координирует работу с репозиториями

3. **Repository Layer** (`internal/repository/`)
   - Инкапсулирует логику доступа к данным
   - Реализует интерфейсы из domain слоя

4. **Handler Layer** (`internal/handler/`)
   - Обрабатывает HTTP запросы
   - Валидирует входные данные

5. **Infrastructure Layer** (`internal/infrastructure/`)
   - Внешние зависимости (БД, логгер)
   - Конфигурация приложения

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|-----------|----------|----------------------|
| `SERVER_ADDRESS` | Адрес HTTP сервера | `:8080` |
| `DATABASE_URL` | URL подключения к PostgreSQL | `postgres://quotes_user:quotes_pass@localhost:5432/quotes_db?sslmode=disable` |
| `LOG_LEVEL` | Уровень логирования | `info` |
| `DB_MAX_OPEN_CONNS` | Максимум открытых соединений | `25` |
| `DB_MAX_IDLE_CONNS` | Максимум idle соединений | `25` |
| `DB_CONN_MAX_LIFETIME` | Время жизни соединения | `5m` |

### Пример .env файла

```bash
SERVER_ADDRESS=:8080
DATABASE_URL=postgres://quotes_user:quotes_pass@postgres:5432/quotes_db?sslmode=disable
LOG_LEVEL=debug
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=10m
```

## 📊 База данных

### Схема таблицы quotes

```sql
CREATE TABLE quotes (
    id SERIAL PRIMARY KEY,
    author VARCHAR(100) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_quotes_author ON quotes(author);
CREATE INDEX idx_quotes_created_at ON quotes(created_at);
```

### Миграции

Миграции выполняются автоматически при старте приложения:

```bash
# Ручной запуск миграций
docker exec -it quotes-service_postgres_1 psql -U quotes_user -d quotes_db -f /docker-entrypoint-initdb.d/001_create_quotes_table.sql
```

## 🛠️ Команды разработки

```bash
# Сборка приложения
make build

# Запуск приложения
make run

# Запуск тестов
make test

# Запуск тестов с покрытием
make test-coverage

# Линтинг кода
make lint

# Форматирование кода
make fmt

# Docker команды
make docker-up          # Запуск всех сервисов
make docker-down        # Остановка всех сервисов
make docker-build       # Пересборка образов
make docker-logs        # Просмотр логов

# Очистка
make clean
```

## 🧪 Тестирование

### Unit тесты

```bash
# Запуск всех тестов
go test ./...

# Тесты с покрытием
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Тесты конкретного пакета
go test ./internal/service/...
```

### Integration тесты

```bash
# Запуск integration тестов (требует запущенную БД)
go test -tags=integration ./tests/integration/...
```

### Пример тестирования API

```bash
# Создание тестовых данных
curl -X POST http://localhost:8080/quotes -d '{"author":"Test Author","quote":"Test Quote"}'

# Проверка получения данных
curl http://localhost:8080/quotes

# Очистка тестовых данных
curl -X DELETE http://localhost:8080/quotes/1
```

## 📈 Мониторинг и логирование

### Structured Logging

Все логи выводятся в JSON формате:

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Quote created successfully",
  "quote_id": 1,
  "author": "Confucius"
}
```

### Health Check

Endpoint `/health` предоставляет информацию о состоянии сервиса:

```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "database": "connected",
    "uptime": "1h23m45s"
  }
}
```

### Metrics (будущее развитие)

Можно добавить Prometheus metrics:
- Количество HTTP запросов
- Время ответа API
- Количество цитат в БД
- Статус подключения к БД

## 🔒 Безопасность

### Implemented

- ✅ **SQL Injection Protection** - параметризованные запросы
- ✅ **Input Validation** - валидация всех входных данных
- ✅ **Error Handling** - безопасная обработка ошибок
- ✅ **Resource Limits** - ограничения на размер полей

### Частые проблемы

**1. Приложение не может подключиться к БД**
```bash
# Проверить статус PostgreSQL
docker-compose ps postgres

# Проверить логи
docker-compose logs postgres

# Проверить подключение вручную
docker exec -it quotes-service_postgres_1 psql -U quotes_user -d quotes_db -c "SELECT 1;"
```

**2. Порт 8080 уже занят**
```bash
# Найти процесс, использующий порт
lsof -i :8080

# Изменить порт в docker-compose.yml или переменной SERVER_ADDRESS
```

**3. Медленные запросы к БД**
```bash
# Проверить индексы
docker exec -it quotes-service_postgres_1 psql -U quotes_user -d quotes_db -c "\d quotes"

# Проверить план выполнения запроса
EXPLAIN ANALYZE SELECT * FROM quotes WHERE author ILIKE '%confucius%';
```

### Логи и отладка

```bash
# Логи приложения
docker-compose logs -f quotes-service

# Логи базы данных
docker-compose logs -f postgres

# Подключение к контейнеру для отладки
docker exec -it quotes-service_quotes-service_1 sh
```
