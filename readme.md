# Websoket

## Задачи

1. Выбрать вариант реализации, на уровне nginx, соединия между websocket сервисом и клиентом
2. Оптимизировать значения таймауты (WebSocket Ping) под связку Docker + Nginx.
3. Использование TLS/SSL (WSS)
4. Мониторинг соединений на уровне Nginx
5. Балансировка
6. Какой размер данных максимально можно передавать через websocket и как этим управлять

## TODO

0. Сжатие данных при передаче тесктовых сообщений (Sec-WebSocket-Extensions: permessage-deflate)
1. Мониторинг Nginx в Prometeus
2. Мониторинг на уровне сервиса Prometeus
4. Авторизация через Keycloak
5. pull модель событий из очереди NATS
5. Определить библиотеку golang для реализации websocket
6. Определить клиента для подключения по websocket

# Nginx

Образ [Docker](https://hub.docker.com/_/nginx) nginx:stable-alpine3.23 содержит последнюю стабильную версию (1.28.1)
веб-сервера Nginx.

## Варианты настройки

1. **njs** (Nginx JavaScript) - это наиболее современный способ. Модуль ngx_http_js_module позволяет перехватывать данные и выполнять асинхронные HTTP-запросы.Как работает: Вы пишете JS-скрипт, который принимает сообщение из WebSocket, парсит его (например, JSON) и инициирует внутренний подзапрос (r.subrequest) или внешний запрос через ngx.fetch к вашему REST API.

2. **OpenResty** / lua-nginx-module. Библиотека lua-resty-websocket является стандартом для таких задач. Как работает: Внутри Lua-обработчика вы читаете фреймы WebSocket, извлекаете команду и используете HTTP-клиент (например, lua-resty-http) для отправки REST-запроса. Это обеспечивает 100% неблокирующее поведение.

3. **Nginx как API Gateway**. Сам Nginx часто используется как [прокси](https://nginx.org/ru/docs/http/websocket.html), который просто «пробрасывает» WebSocket до бэкенда (Go), где и происходит основная логика преобразования.

Выбираем третий вариант

# Server

Golang версия 1.25.6

## Выбор библиотеки

- [gorilla/websocket](https://github.com/gorilla/websocket/)
- [coder/websocket](https://github.com/coder/websocket)

Для работы с WebSocket используем библиотеку [gorilla/websocket](https://github.com/gorilla/websocket/)

# Client

```
$ sudo wget -qO ~/.local/bin/websocat https://github.com/vi/websocat/releases/latest/download/websocat.x86_64-unknown-linux-musl
```

[wscat](https://www.npmjs.com/package/wscat) - утилита командной строки на базе Node.js, используемая для тестирования и отладки WebSocket-серверов.

## Запуск

**Вариант 1**. Отображение уведомлений: По умолчанию wscat не показывает скрытые управляющие фреймы. Чтобы увидеть, когда приходят ping или pong, используйте флаг:

```
$ wscat -c ws://localhost/ws-notifications -P
```

**Вариант 2**. Ручная отправка: Если вы хотите отправить ping или pong самостоятельно (например, для отладки), включите режим «slash-команд».После этого можно вводить /ping [сообщение] или /pong [сообщение] прямо в консоли.

```
$ wscat -c ws://localhost/ws-notifications --slash
```

**Вариант 3**. При использовании самоподписанного сертификата wscat откажется подключаться. Используйте флаг -n (no-check), чтобы игнорировать проверку безопасности:

```
$ wscat -n -c wss://localhost/ws-notifications -P
```

# Реализация

**Шаг 1**. Чтобы Nginx вообще понимал WebSocket, необходимо явно передать заголовки Upgrade и Connection.

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

**Шаг 2** Для WebSocket это самая важная настройка. Если она включена (on), Nginx будет пытаться накопить данные от вашего Go-сервера в буфер, прежде чем отправить их клиенту. Это создаст огромные задержки («лаги») в стриме данных.

```
location /ws-notifications/ {
    ...
    # Отключаем буферизацию, чтобы сообщения улетали мгновенно
    proxy_buffering off;
    ...
}
```

**Шаг 3**. По умолчанию соединение будет закрыто, если с проксируемого сервера данные не передавались в течение 60 секунд. Этот таймаут можно увеличить при помощи директивы конфигурационного файла Nginx [proxy_read_timeout](https://nginx.org/ru/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). Кроме того, на проксируемом сервере можно настроить периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.

> [!NOTE]
>Nginx должен иметь proxy_read_timeout больше, чем ваш pongWait, иначе он убьет соединение раньше, чем сработает логика в Go.

```
location /ws-notifications/ {
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

На стороне сервиса определяем периодическую отправку WebSocket ping-фреймов для сброса таймаута и проверки работоспособности соединения.

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

**Шаг 4**. Данные передаются как текст в кодировке UTF-8 (websocket.TextMessage = 1). В пакете [github.com/gorilla/websocket](https://github.com/gorilla/websocket) определено несколько типов сообщений, которые соответствуют спецификации протокола RFC 6455.

**Шаг 5**. Без установленного дедлайна WriteMessage может зависнуть навсегда, если клиент перестал подтверждать получение пакетов, а системные буферы переполнились. Разработчики библиотеки [github.com/gorilla/websocket](https://github.com/gorilla/websocket) настоятельно рекомендуют всегда устанавливать дедлайн перед каждой операцией записи, чтобы избежать утечки горутин.

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

**Шаг 6**. Для локальной разработки или использования внутри закрытой сети (Docker-сеть) проще всего выпустить самоподписанный сертификат (self-signed). Если вы хотите, чтобы браузер или wscat не ругались на недоверенный сертификат, используйте инструмент mkcert. Он создает локальный центр сертификации. Вы получите два файла: localhost.pem и localhost-key.pem

```
$ mkcert localhost 127.0.0.1 172.18.0.1
```

Настроить WSS (WebSocket Secure) в Nginx.

```
server {
    listen 443 ssl;
    server_name localhost;

    ssl_certificate     /etc/nginx/localhost.pem;
    ssl_certificate_key /etc/nginx/localhost-key.pem;
}
```

**Шаг 7** Мониторинг соединений на уровне Nginx. Можно смотреть количество активных соединений через модуль stub_status.

```
location /status {
    stub_status;
}
```

Команда curl покажет строку Active connections, куда входят вебсокеты

```
$ curl -k  https://localhost/status
Active connections: 1
```

**Шаг 8** Балансировка WebSocket сложнее обычной, так как это stateful протокол: соединение держится долго, и разорвать его — значит заставить клиента переподключаться.

```
upstream go_sockets {
    # Sticky sessions важны, если сервер хранит локальный стейт клиента
    ip_hash;

    server wsserver:8000;
}
```

Запускаем идентичные копии сервиса (например, 3 воркера), используем флаг --scale

```
$ docker-compose up --remove-orphans -d --scale wsserver=3
```

> [!NOTE]
> в docker-compose.yaml у этого сервиса не должно быть жестко прописано container_name. Docker сам назначит имена (например, websocket-wsserver-1, websocket-wsserver-2, websocket-wsserver-3)

Чтобы прописать масштабируемые сервисы в Nginx, вам не нужно знать индивидуальные имена контейнеров. Внутренний DNS Docker автоматически связывает имя сервиса (в вашем случае wsserver) со всеми его IP-адресами. Когда Nginx обращается к хосту wsserver, DNS-сервер Docker возвращает список IP-адресов всех запущенных реплик этого сервиса.

**Шаг 9** Теоретически предел одного фрейма в протоколе WebSocket — 2^63 байт (9 эксабайт), но на практике вы упретесь в лимиты библиотеки, памяти и прокси-сервера.

> [!NOTE]
> Золотое правило:
> * Для текста/JSON: не более 1–5 МБ. Если больше — парсинг JSON заблокирует поток.
> * Для бинарных данных: до 10–50 МБ. Если нужно передавать файлы больше, лучше использовать чанки (куски) или просто отдавать ссылку на скачивание по HTTP.

По умолчанию библиотека [github.com/gorilla/websocket](https://github.com/gorilla/websocket) не ограничивает размер сообщения, но для защиты от DoS-атак (переполнения памяти) рекомендуется устанавливать лимит:

```
const (
    ...
    // ограничивает размер сообщения,
    // для защиты от DoS-атак (переполнения памяти)
    readLimit = 65536 // Ограничение на 64 КБ
    ...
)

func reader(ws *websocket.Conn, echo chan string) {
    ...
	SetReadLimit(readLimit)
    ...
}
```

По умолчанию в Nginx стоит лимит 1 МБ. Если вы планируете передавать через сокет тяжелые файлы или большие пачки данных в одном сообщении, этот лимит нужно снять или увеличить. Даже если Go проглотит гигабайт, Nginx может его обрубить.

```
location /ws-notifications {
    ...
    # Максимальный размер тела запроса (0 = без ограничений).
    # Помогает, если передаются очень большие начальные данные.
    client_max_body_size 0;

    # Размер буфера для чтения тела запроса.
    # Если данные больше 128k, Nginx запишет их во временный файл.
    client_body_buffer_size 128k;
}
```

> [!NOTE]
> Если вы поставили в Go ws.SetReadLimit(65536), то настройки Nginx выше не имеют практического смысла, так как Go сам «отрубит» клиента раньше, чем Nginx заметит перегрузку.

Чтобы отправить сообщение размером более 64 КБ через websocat, его нужно передать из файла. Попытка вставить такой объем текста вручную в терминал обычно обрезается лимитами самого эмулятора терминала или оболочки (shell).

```
# Создаем тестовый файл размером 100КБ
$ base64 /dev/urandom | tr -d '\n' | head -c 102400 > big_message.txt

# Отправляем файл в сокет
$ cat big_message.txt | websocat -v -k -B 200000 wss://localhost/ws-notifications
```
