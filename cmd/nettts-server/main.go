package main

import (
	"encoding/binary"
	"net"

	"github.com/tzneal/gopicotts"
)

func main() {
	listen, _ := net.Listen("tcp", "0.0.0.0:5555")
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
	ttsEngine, _ := gopicotts.NewEngine(gopicotts.DefaultOptions)
	defer ttsEngine.Close()
	defer conn.Close()
	ttsEngine.SetOutput(func(c []int16) {
		binary.Write(conn, binary.LittleEndian, c)
	})
	for {
		buf := make([]byte, 8192)
		n, e := conn.Read(buf)
		if e != nil {
			break
		}
		ttsEngine.SendText(string(buf[:n]))
		ttsEngine.FlushSendText()
	}
}
