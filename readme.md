# Websoket

# Nginx

Образ (Docker)[https://hub.docker.com/_/nginx] nginx:stable-alpine3.23 содержит последнюю стабильную версию (1.28.1)
веб-сервера Nginx.

## Develop websocket server

Golang версия 1.25.6

### Разработка в контейнере

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

### Тестирование 

```
$ wscat -c ws://localhost:8000/ws-notifications
Connected (press CTRL+C to quit)
> hello
< hello

```
