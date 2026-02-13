package main

import (
    "os"
    "fmt"
    "log"
    "context"
    "net/http"
    "encoding/json"
    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

// App содержит зависимости нашего сервера
type App struct {
	js jetstream.JetStream
}

//
type Notification struct {
    ID      string `json:"id"`
    Message string `json:"message"`
}

// sseHandler отдельный метод структуры App
func (app *App) sseHandler(w http.ResponseWriter, r *http.Request) {
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

	// Создаем consumer для этого клиента
	// Он будет слушать общие события и персональные
    ctx := r.Context()
	cons, err := app.js.CreateOrUpdateConsumer(ctx, "notifications", jetstream.ConsumerConfig{
		FilterSubjects: []string{"events.broadcast", "events.user." + userID},
		DeliverPolicy:  jetstream.DeliverNewPolicy, // Только новые уведомления
	})
	if err != nil {
		log.Printf("Consumer error: %v", err)
		return
	}

	// Потребляем сообщения
	iter, _ := cons.Messages()
	defer iter.Stop()

    log.Printf("Пользователь %s подключен\n", userID)

	// Цикл отправки событий
	for {
		select {
		case <-r.Context().Done():
			// Если клиент закрыл вкладку/соединение
			log.Printf("Пользователь %s отключился\n", userID)
			return
		default:
			msg, err := iter.Next()
			if err != nil {
				continue
			}

			// Отправка в SSE формат
			fmt.Fprintf(w, "data: %s\n\n", string(msg.Data()))
			flusher.Flush()

			msg.Ack() // Подтверждаем получение
		}
	}
}

//
func (app *App) startHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id") // В проде берем из JWT
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
    log.Printf("Пользователь %s запустил процесс\n", userID)

    notification, _ := json.Marshal(Notification{ID: "1", Message: "Hello!"})
    subject := fmt.Sprintf("events.user.%s", userID)

	// Отправляем персонально пользователю
    ctx := context.Background()
	ack, err := app.js.Publish(ctx, subject, []byte(notification))
	if err != nil {
		log.Fatal("Ошибка публикации в %s: %v", subject, err)
		http.Error(w, "Ошибка публикации", http.StatusInternalServerError)
		return
	}
	// Печатаем ID сообщения в стриме, подтверждение доставки в NATS
	log.Printf("Сообщение доставлено в NATS! Stream: %s, Sequence: %d\n", ack.Stream, ack.Sequence)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    w.Write([]byte(notification))
}

//
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

    // Инициализация JetStream
	js, _ := jetstream.New(nc)
    app := &App{js: js}

	// Создаем стрим для уведомлений
    ctx := context.Background()
	cfg := jetstream.StreamConfig{
		Name:     "notifications",
		Subjects: []string{"events.>"}, // Слушаем все в этом пространстве
	}
	app.js.CreateOrUpdateStream(ctx, cfg)

	http.HandleFunc("/sse-notifications", app.sseHandler)
	http.HandleFunc("/start-process", app.startHandler)

	log.Println("Сервер запущен на :80")
    err = http.ListenAndServe(":80", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

