

# Ограничения на размер сообщения

Формально в спецификации SSE нет жесткого лимита на размер одного сообщения. Однако на практике есть ограничения инфраструктуры.

**Буферы Nginx**

Если сообщение больше, чем proxy_buffer_size и буферизация включена, данные будут копиться.

**Лимиты NATS**

По умолчанию максимальный размер сообщения в NATS — 1 МБ.
Его можно увеличить в конфиге сервера, но NATS не предназначен для передачи тяжелых файлов. Оптимально держать сообщения в пределах 64 КБ – 256 КБ.

**Память браузера**

Браузер парсит текстовый поток. Если вы пришлете JSON на 100 МБ, JSON.parse может заморозить поток выполнения или вызвать Out of Memory.

# TypeSpec

```
npm install -g @typespec/compiler
```

```
$ npm install @typespec/http
$ npm install @typespec/openapi3
```

Для генерации кода на Go из TypeSpec используется официальный эмиттер @typespec/http-client-csharp (на текущий момент официальный эмиттер для Go находится в стадии разработки/активного сообщества) или более универсальный путь через OpenAPI.

Шаг 1: Генерация OpenAPI спецификации

Компилируем TypeSpec в JSON

```
$ tsp compile notifications.tsp --emit @typespec/openapi3
```

После этого в папке tsp-output/@typespec/openapi3 появится файл openapi.yaml

Шаг 3: Генерация Go-структур

Используем полученный файл для генерации структур

```
$ oapi-codegen -generate types -package models tsp-output/@typespec/openapi3/openapi.yaml > models.gen.go
```

Альтернативно можно использовать TypeSpec-Go (Experimental)

> [!NOTE]
> Может поддерживать не все функции TypeSpec v1.9.0

Установить

```
$ npm install @typespec/http-client-go
```

Запустить

```
$ tsp compile notifications.tsp --emit @typespec/http-client-go
```

# Swagger UI

Вот прямая ссылка на страницу со всеми [релизами Swagger UI](https://github.com/swagger-api/swagger-ui/releases)
После распаковки понадобится только содержимое папки dist


# Тестирование

```
$ curl -k -v http://localhost/sse-notifications?user_id=rootcom
```

```
$ curl -k -v http://localhost/start-process?user_id=rootcom | jq
```


