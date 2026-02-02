# Websoket



## Develop websocket server

Golang версия 1.25.6

### Разработка в контейнере

1. Собрать контейнер

```
$ docker-compose -f docker-compose-dev.yaml up
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
$ wscat -c ws://localhost:8000/ws
Connected (press CTRL+C to quit)
> hello
< hello

```
