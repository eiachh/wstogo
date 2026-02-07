package main

import (
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.RWMutex
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()
	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()
		c.Close()
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// broadcast time every sec to all clients
func broadcaster() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		now := time.Now().Format(time.RFC3339)
		clientsMu.RLock()
		for c := range clients {
			c.WriteMessage(websocket.TextMessage, []byte(now))
		}
		clientsMu.RUnlock()
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	go broadcaster()
	http.HandleFunc("/echo", echo)
	log.Printf("echo server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
