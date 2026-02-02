# Develop websocket server

## Разработка в контейнере

1. Собрать контейнер

```
$ docker-compose -f docker-compose-dev.yaml up --remove-orphans
```

2. Подключиться в режиме командной строки

```
$ docker exec -it ws bash
```

3. Запуск сервиса

```
# go run main.go
```

## Тестирование в контейнере

```
$ wscat -c ws://localhost:8000/ws-notifications
Connected (press CTRL+C to quit)
> hello
< hello
```

## Сервисное тестирование

```
$ docker-compose up --remove-orphans
```

```
$ wscat -c ws://localhost/ws-notifications
Connected (press CTRL+C to quit)
> hello
< hello

```

[] Timeout - не активное соединение должно отваливаться через 120 сек


