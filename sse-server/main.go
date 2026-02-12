package main

import (
    "os"
    "fmt"
	"log"
	"time"
    //"context"
	"net/http"
    "github.com/nats-io/nats.go"
    //"github.com/nats-io/nats.go/jetstream"
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id") // В проде берем из JWT
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Устанавливаем обязательные заголовки
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("Access-Control-Allow-Origin", "*") // Для CORS, если нужно

	// Получаем Flusher, чтобы отправлять данные немедленно
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	log.Println("Клиент подключился")

	// Цикл отправки событий
	for {
		select {
		case <-r.Context().Done():
			// Если клиент закрыл вкладку/соединение
			log.Println("Клиент отключился")
			return
		default:
			// Формат SSE: "data: <сообщение>\n\n"
			currentTime := time.Now().Format("15:04:05")
			fmt.Fprintf(w, "data: Текущее время сервера: %s\n\n", currentTime)

			// Сбрасываем буфер в сеть прямо сейчас
			flusher.Flush()

			// Ждем 2 секунды перед следующим событием
			time.Sleep(2 * time.Second)
		}
	}
}

func main() {
	// Подключение к NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
    log.Println(nc)

	http.HandleFunc("/sse-notifications", sseHandler)

	log.Println("Сервер запущен на :80")
    err = http.ListenAndServe(":80", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

