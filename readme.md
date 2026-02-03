# Websoket

## Задачи

1. Определить библиотеку golang для реализации websocket
2. Определить клинта для подключения по websocket
3. Выбрать вариант реализации, на уровне nginx, соединия между websocket сервисом и клиентом
4. Оптимизировать значения таймауты (WebSocket Ping) под связку Docker + Nginx.

TODO

1. Мониторинг соединений на уровне Nginx
2. Использование TLS/SSL (WSS)
3. Балансировка нагрузки
4. Мониторинг на уровне сервиса

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

        location /ws-notifications/ {
            proxy_pass http://backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
            ...
        }
    }
}
```

2. По умолчанию соединение будет закрыто, если с проксируемого сервера данные не передавались в течение 60 секунд. Этот таймаут можно увеличить при помощи директивы конфигурационного файла Nginx [proxy_read_timeout](https://nginx.org/ru/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). Кроме того, на проксируемом сервере можно настроить периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.

> [!NOTE]
>Nginx должен иметь proxy_read_timeout больше, чем ваш pongWait, иначе он убьет соединение раньше, чем сработает логика в Go.

```
location /ws-notification/ {
    ...
    # Запас в Nginx (90с) гарантирует, что если пакет задержится
    # на 5-10 секунд, прокси не обрубит связь "кодом 1006"
    proxy_read_timeout 90s;
    proxy_send_timeout 90s;

    # Отключаем буферизацию, чтобы сообщения улетали мгновенно
    proxy_buffering off;
    ...
}
```

# Server

Golang версия 1.25.6

Для работы с WebSocket используем библиотеку [gorilla/websocket](https://github.com/gorilla/websocket/)

## Реализация

1. Периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.

```
const (
    // Время на отправку сообщения
    // Дает больше времени сетевому стеку на очистку буферов при временных затыках.
	writeWait = 15 * time.Second        // 15 сек
    // Сколько ждем ответа от клиента
	pongWait = 60 * time.Second         // 60 сек
    // Пингуем чаще, чтобы Nginx не закрыл сокет
    // Чем чаще пинг, тем быстрее вы узнаете о разрыве
	pingPeriod = 30 * time.Second       // 30 сек
)
```

2. Данные передаются как текст в кодировке UTF-8 (websocket.TextMessage = 1). В пакете [github.com/gorilla/websocket](https://github.com/gorilla/websocket) определено несколько типов сообщений, которые соответствуют спецификации протокола RFC 6455.

3. Без установленного дедлайна WriteMessage может зависнуть навсегда, если клиент перестал подтверждать получение пакетов, а системные буферы переполнились. Разработчики библиотеки [github.com/gorilla/websocket](https://github.com/gorilla/websocket) настоятельно рекомендуют всегда устанавливать дедлайн перед каждой операцией записи, чтобы избежать утечки горутин.

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

