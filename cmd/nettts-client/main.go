package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"os/signal"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

func main() {
	addr := flag.String("ip", "ws://nettts.mserv.kab/ws", "IP:port of the server")
	flag.Parse()
	conn, _, e := websocket.DefaultDialer.Dial(*addr, http.Header{})
	if e != nil {
		panic(e.Error())
	}
	defer conn.Close()
	println("Connection opened")
	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 1)
	stream, err := portaudio.OpenDefaultStream(0, 1, 16000, len(out), &out)
	if err != nil {
		panic(err)
	}
	err = stream.Start()
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	brcon := bufio.NewReaderSize(conn.UnderlyingConn(), 65536)
	bwcon := bufio.NewWriterSize(conn.UnderlyingConn(), 65536)
	if err != nil {
		println(err.Error())
	}
	trm := make(chan os.Signal, 4)
	signal.Notify(trm, os.Interrupt)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			buf, _ := reader.ReadString('\n')
			err := ttsserver.SendTTSRequest(bwcon, buf)
			if err != nil {
				println("SEND ERROR " + err.Error())
				break
			}
			err = bwcon.Flush()
			if err != nil {
				println("FLUSH ERROR" + err.Error())
				break
			}
		}
	}()
	go func() {
		for {
			resp, err := ttsserver.RecvTTSResponse(brcon)
			if err != nil {
				println(err.Error())
				break
			}
			buf := resp.TTSData
			for len(buf) > 0 {
				out[0] = buf[0]
				stream.Write()
				buf = buf[1:]
			}
		}
	}()
	<-trm
}
