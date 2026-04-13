package podapi

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Allow all origins for development; restrict this in production.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketTest(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		var t terminalSession

		t.inputchan <- message
		log.Printf("Received: %s", message)

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

type terminalSession struct {
	Wscon     *websocket.Conn
	inputchan chan []byte
}

func (t *terminalSession) Read(p []byte) (int, error) {
	data, ok := <-t.inputchan
	if !ok {
		return 0, io.EOF
	}
	n := copy(p, data)
	return n, nil

}
