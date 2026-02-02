package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024, // Размер буфера чтения
    WriteBufferSize: 1024, // Размер буфера записи
    // Позволяет определить, должен ли сервер сжимать сообщения
    EnableCompression: true,
}

func reader(ws *websocket.Conn) {
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

func handleConnections(w http.ResponseWriter, r *http.Request) {
    // обновление соединения до WebSocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    reader(ws)
}

func main() {
    http.HandleFunc("/ws-notifications", handleConnections)
    log.Println("http server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

