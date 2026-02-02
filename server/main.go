package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    // обновление соединения до WebSocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // цикл обработки сообщений
    for {
        messageType, message, err := ws.ReadMessage()
        if err != nil {
            log.Println(err)
            break
        }
        log.Printf("Received: %s", message)

        // эхо ансвер
        if err := ws.WriteMessage(messageType, message); err != nil {
            log.Println(err)
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleConnections)
    log.Println("http server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

