package main

import (
	"bufio"
	"log"
	"net"
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
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	handleConn(ws.UnderlyingConn())
}

func main() {
	http.HandleFunc("/ws", wsEndpoint)
	http.ListenAndServe(":5555", nil)
}

func handleConn(conn net.Conn) {
	println("New Connection")
	ttsEngine, err := gopicotts.NewEngine(gopicotts.DefaultOptions)
	defer conn.Close()
	defer ttsEngine.Close()
	if err != nil {
		println(err.Error())
		return
	}
	bwcon := bufio.NewWriterSize(conn, 65536)
	brcon := bufio.NewReaderSize(conn, 65536)
	ttsEngine.SetOutput(func(c []int16) {
		err := ttsserver.SendTTSResponse(bwcon, c)
		if err != nil {
			println("SEND ERROR " + err.Error())
		}
		err = bwcon.Flush()
		if err != nil {
			println("FLUSH ERROR" + err.Error())
		}
	})
	for {
		req, err := ttsserver.RecvTTSRequest(brcon)
		if err != nil {
			println(err.Error())
			break
		}
		ttsEngine.SendText(req.Request)
		ttsEngine.FlushSendText()
	}
	println("Connection Closed!")
}
