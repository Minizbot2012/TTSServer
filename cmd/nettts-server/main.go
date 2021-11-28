package main

import (
	"bufio"
	"net"
	"time"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/tzneal/gopicotts"
)

func main() {
	listen, err := net.Listen("tcp", "0.0.0.0:5555")
	if err != nil {
		panic(err.Error())
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			println(err.Error())
		} else {
			go handleConn(conn)
		}
	}
}

func handleConn(conn net.Conn) {
	println("New Connection")
	ttsEngine, err := gopicotts.NewEngine(gopicotts.DefaultOptions)
	if err != nil {
		println(err.Error())
		conn.Close()
		return
	}
	conn.SetDeadline(time.Time{})
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
		Req, err := ttsserver.RecvTTSRequest(brcon)
		if err != nil {
			println(err.Error())
			break
		}
		ttsEngine.SendText(Req.Request)
		ttsEngine.FlushSendText()
	}
	conn.Close()
	ttsEngine.Close()
	println("Connection Closed!")
}
