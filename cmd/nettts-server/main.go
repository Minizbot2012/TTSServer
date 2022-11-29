package main

import (
	"log"
	"net/http"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/gorilla/websocket"
	"github.com/tzneal/gopicotts"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	handleConn(ws)
}

func main() {
	http.HandleFunc("/ws", wsEndpoint)
	http.ListenAndServe(":5555", nil)
}

func handleConn(conn *websocket.Conn) {
	defer conn.Close()
	ttsEngine, err := gopicotts.NewEngine(gopicotts.DefaultOptions)
	if err != nil {
		return
	}
	defer ttsEngine.Close()
	conn.SetPingHandler(func(appData string) error {
		e := conn.WriteMessage(websocket.PongMessage, []byte(appData))
		return e
	})
	ttsEngine.SetOutput(func(c []int16) {
		err := ttsserver.SendTTSResponse(conn, c)
		if err != nil {
			println("SEND ERROR " + err.Error())
		}
	})
	for {
		req, err := ttsserver.RecvTTSRequest(conn)
		if err != nil {
			break
		}
		ttsEngine.SendText(req.Request)
		ttsEngine.FlushSendText()
	}
	println("Connection Closed!")
}
