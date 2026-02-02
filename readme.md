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

## Реализация

1. Периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.
2. Данные передаются как текст в кодировке UTF-8 (websocket.TextMessage = 1). В пакете [github.com/gorilla/websocket](https://github.com/gorilla/websocket) определено несколько типов сообщений, которые соответствуют спецификации протокола RFC 6455.
3. Без установленного дедлайна ваш WriteMessage может «зависнуть» навсегда, если клиент перестал подтверждать получение пакетов, а системные буферы переполнились. Разработчики библиотеки [github.com/gorilla/websocket](https://github.com/gorilla/websocket) настоятельно рекомендуют всегда устанавливать дедлайн перед каждой операцией записи, чтобы избежать утечки горутин.

```
    for {
        select {
            case message := <-echo:
                ws.SetWriteDeadline(time.Now().Add(writeWait))
                if err := ws.WriteMessage(websocket.TextMessage,[]byte(message)); err != nil {
                    log.Println("EchoWriteMessageError:", err)
                    return
                }
            case <-pingTicker.C:
                ws.SetWriteDeadline(time.Now().Add(writeWait))
                if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
                    log.Println("PingWriteMessageError:", err)
                    return
                }
        }
    }
```

# Client

[wscat](https://www.npmjs.com/package/wscat) - утилита командной строки на базе Node.js, используемая для тестирования и отладки WebSocket-серверов.

## Команды для работы с ping/pong

1. Отображение уведомлений: По умолчанию wscat не показывает скрытые управляющие фреймы. Чтобы увидеть, когда приходят ping или pong, используйте флаг:

```
wscat -c <url> -P
```

2. Ручная отправка: Если вы хотите отправить ping или pong самостоятельно (например, для отладки), включите режим «slash-команд».После этого можно вводить /ping [сообщение] или /pong [сообщение] прямо в консоли.

```
wscat -c <url> --slash
```

