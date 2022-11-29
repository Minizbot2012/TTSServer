package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	if err != nil {
		println(err.Error())
	}
	keepAlive(conn, time.Second*10)
	trm := make(chan os.Signal, 4)
	signal.Notify(trm, os.Interrupt)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			buf, _ := reader.ReadString('\n')
			err := ttsserver.SendTTSRequest(conn, buf)
			if err != nil {
				println("SEND ERROR " + err.Error())
				break
			}
			if err != nil {
				println("FLUSH ERROR" + err.Error())
				break
			}
		}
	}()
	go func() {
		for {
			resp, err := ttsserver.RecvTTSResponse(conn)
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

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				return
			}
			time.Sleep(timeout / 2)
			if time.Since(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}
