# Websoket

# Nginx

Образ [Docker](https://hub.docker.com/_/nginx) nginx:stable-alpine3.23 содержит последнюю стабильную версию (1.28.1)
веб-сервера Nginx.

## Варианты настройки

1. **njs** (Nginx JavaScript) - это наиболее современный способ. Модуль ngx_http_js_module позволяет перехватывать данные и выполнять асинхронные HTTP-запросы.Как работает: Вы пишете JS-скрипт, который принимает сообщение из WebSocket, парсит его (например, JSON) и инициирует внутренний подзапрос (r.subrequest) или внешний запрос через ngx.fetch к вашему REST API.

2. **OpenResty** / lua-nginx-module. Библиотека lua-resty-websocket является стандартом для таких задач. Как работает: Внутри Lua-обработчика вы читаете фреймы WebSocket, извлекаете команду и используете HTTP-клиент (например, lua-resty-http) для отправки REST-запроса. Это обеспечивает 100% неблокирующее поведение.

3. **Nginx как API Gateway**. Сам Nginx часто используется как [прокси](https://nginx.org/ru/docs/http/websocket.html), который просто «пробрасывает» WebSocket до бэкенда (Go), где и происходит основная логика преобразования.

## Настройка проксирования Nginx как API Gateway

1. Чтобы Nginx вообще понимал WebSocket, необходимо явно передать заголовки Upgrade и Connection.

```
http {
    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }

    server {
        ...

        location /chat/ {
            proxy_pass http://backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }
    }
}
```

2. По умолчанию соединение будет закрыто, если с проксируемого сервера данные не передавались в течение 60 секунд. Этот таймаут можно увеличить при помощи директивы конфигурационного файла Nginx [proxy_read_timeout](https://nginx.org/ru/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). Кроме того, на проксируемом сервере можно настроить периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.

```
proxy_read_timeout 120s
```

# Server

Golang версия 1.25.6

Для работы с WebSocket используем библиотеку [gorilla/websocket](https://github.com/gorilla/websocket/)

## Develop websocket server


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

### Тестирование в контейнере

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


