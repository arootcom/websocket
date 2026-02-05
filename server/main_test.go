package main

import (
    "fmt"
    "log"
    "testing"
    "strings"
    "net/http"
    "crypto/rand"
    "encoding/base64"
    "net/http/httptest"
    "github.com/gorilla/websocket"
)

func TestCompressionWebSocketHandler(t *testing.T) {
	// 1. Создаем тестовый сервер
	s := httptest.NewServer(http.HandlerFunc(handleConnections))
	defer s.Close()

	// 2. Преобразуем http:// в ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

    dialer := websocket.Dialer{
        EnableCompression: true,
    }

	// 3. Подключаемся клиентом
    ws, resp, err := dialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Не удалось подключиться: %v", err)
	}
	defer ws.Close()

    // 4. Проверяем, что сервер подтвердил сжатие в заголовках
    ext := resp.Header.Get("Sec-WebSocket-Extensions")
    if !strings.Contains(ext, "permessage-deflate") {
        t.Error("Сжатие не согласовано сервером")
    }

	// 5. Отправляем сообщение
	input := []byte("hello")
	if err := ws.WriteMessage(websocket.TextMessage, input); err != nil {
		t.Fatalf("Ошибка записи: %v", err)
	}

	// 6. Читаем ответ (если сервер делает эхо)
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Ошибка чтения: %v", err)
	}

	if string(message) != string(input) {
		t.Errorf("Ожидалось %s, получили %s", input, message)
	}

    // 7. Проверяем, что сервер реально обрывает связь на 64 КБ
	size := 102400 // 100 КБ

	// Генерируем случайные байты (аналог /dev/urandom)
	// Важно: берем чуть меньше байт, так как base64 увеличивает размер на ~33%
	rawBytes := make([]byte, size)
	_, err = rand.Read(rawBytes)
	if err != nil {
		log.Fatal(err)
	}

	// Кодируем в Base64 (стандартный кодировщик в Go не вставляет \n)
	encoded := base64.StdEncoding.EncodeToString(rawBytes)

	// Обрезаем строго до нужного количества символов (аналог head -c)
	if len(encoded) > size {
		encoded = encoded[:size]
	}

	fmt.Printf("Размер: %d байт\n", len(encoded))

    ws.WriteMessage(websocket.TextMessage, []byte(encoded))

    _, _, err = ws.ReadMessage()
    if err == nil {
        t.Error("Сервер должен был закрыть соединение из-за лимита")
    }
}

func TestWebSocketHandler(t *testing.T) {
	// 1. Создаем тестовый сервер
	s := httptest.NewServer(http.HandlerFunc(handleConnections))
	defer s.Close()

	// 2. Преобразуем http:// в ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http")

    dialer := websocket.Dialer{}

	// 3. Подключаемся клиентом
    ws, _, err := dialer.Dial(u, nil)
	if err == nil {
		t.Error("Удалось подключиться")
	    ws.Close()
	}
}
