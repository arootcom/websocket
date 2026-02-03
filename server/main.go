package main

import (
    "log"
    "time"
    "net/http"
    "github.com/gorilla/websocket"
)

const (
    // Время на отправку сообщения
	writeWait = 15 * time.Second        // 15 сек
    // Сколько ждем ответа от клиента
	pongWait = 60 * time.Second         // 60 сек
    // Пингуем чаще, чтобы Nginx не закрыл сокет
    // Чем чаще пинг, тем быстрее вы узнаете о разрыве
	pingPeriod = 30 * time.Second       // 30 сек
)

var (
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024, // Размер буфера чтения
        WriteBufferSize: 1024, // Размер буфера записи
        // Позволяет определить, должен ли сервер сжимать сообщения
        EnableCompression: true,
    }
)

func reader(ws *websocket.Conn, echo chan string) {
    defer ws.Close()

    ws.SetPongHandler( func(appData string) error {
        log.Printf("Получен PONG от %s: %s", ws.RemoteAddr(), appData)
        ws.SetReadDeadline(time.Now().Add(pongWait));
        return nil
    })

    // цикл обработки сообщений
    for {
        _, message, err := ws.ReadMessage()
        if err != nil {
            log.Println(err)
            break
        }
        log.Printf("Received: %s", message)
        echo <- string(message)
    }
}

func writer(ws *websocket.Conn, echo chan string) {
    pingTicker := time.NewTicker(pingPeriod)
    defer pingTicker.Stop()

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
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    // обновление соединения до WebSocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    echo := make(chan string)

    go writer(ws, echo)
    reader(ws, echo)
}

func main() {
    http.HandleFunc("/ws-notifications", handleConnections)
    log.Println("http server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

