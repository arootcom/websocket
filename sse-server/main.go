package main

import (
    "fmt"
	"log"
	"net/http"
	"time"
)

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Устанавливаем обязательные заголовки
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("Access-Control-Allow-Origin", "*") // Для CORS, если нужно

	// 2. Получаем Flusher, чтобы отправлять данные немедленно
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	log.Println("Клиент подключился")

	// 3. Цикл отправки событий
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

			// 4. Сбрасываем буфер в сеть прямо сейчас
			flusher.Flush()

			// Ждем 2 секунды перед следующим событием
			time.Sleep(2 * time.Second)
		}
	}
}

func main() {
	http.HandleFunc("/sse-notifications", sseHandler)

	log.Println("Сервер запущен на :80")
    err := http.ListenAndServe(":80", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

