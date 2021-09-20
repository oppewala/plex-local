package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var videoStatus = make(chan *DownloadTracker)

var upgrader = websocket.Upgrader{
	// TODO: Fix origin header check
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsStatus(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("received: %s", message)

		select {
		case vs := <-videoStatus:
			res := fmt.Sprintf("Video status: %s", vs)
			err = c.WriteMessage(mt, []byte(res))
			if err != nil {
				log.Println("write:", err)
			}
		default:
			continue
		}
	}
}
