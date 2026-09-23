# В Makefile находятся следующие команды:
- make build: Генерация контейнеров Докера и запуск программы с сервером
- make start/make stop: Запуск или выключение контейнеров программы или сервера в Докере
- make start_db/make stop_db: Запуск или выключение контейнера сервера в Докере
- make start_app/make stop_app: Запуск или выключение контейнера программы в Докере
- make remove: Удаление созданных контейнеров в Докере

# Примеры запросов для тестирования приложения:
## Создание успешной записи tasks:
```
POST localhost:8080/tasks
Content-Type: application/json

{
  "Title": "test"
}
```

## Создание ошибочной записи tasks (статус будет выставлен как "failed"):
```
POST localhost:8080/tasks
Content-Type: application/json

{
  "Title": "fail"
}
```

## Выбор всех записей из таблицы tasks:
```GET localhost:8080/tasks```

## Выбор только записей из таблицы tasks у которых стоит статус "failed":
```GET localhost:8080/tasks?status=failed```

## Выбор записи из таблицы tasks с ID 3:
```GET localhost:8080/tasks/3```

## Удаление записи записи из таблицы tasks с ID 3:
```DELETE localhost:8080/tasks/3```

## Health-check:
```DELETE localhost:8080/health```