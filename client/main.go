package main

import (
	"os"
	"log"
	"time"
	"net/url"
	"os/signal"
    "crypto/tls"
    "github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: os.Getenv("WS_SCHEME"), Host: os.Getenv("WS_HOSTNAME"), Path: "/ws-notifications"}
	log.Printf("Подключение к %s", u.String())

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Игнорировать недоверенные сертификаты
	}

	dialer := websocket.Dialer{
		EnableCompression: true,
        TLSClientConfig:   tlsConfig,
		HandshakeTimeout:  10 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer conn.Close()

	// Канал для отслеживания прерывания (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Горутина для чтения ответов от сервера
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка чтения:", err)
				return
			}
			log.Printf("Ответ от сервера: %s", message)
		}
	}()

	// Основной цикл: отправка сообщений каждые 5 секунд
	ticker := time.NewTicker(95 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			// Отправляем сообщение
			msg := []byte("Привет! Текущее время: " + t.String())
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Ошибка записи:", err)
				return
			}
		case <-interrupt:
			// Вежливое закрытие при нажатии Ctrl+C
			log.Println("Завершение работы...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Ошибка при закрытии:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

