package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

func main() {
	listen, err := net.ListenPacket("udp", ":5555")
	if err != nil {
		panic(err)
	}
	ip := flag.String("ip", "ws://nettts.mserv.kab/ws", "IP:port of the server")
	flag.Parse()
	err = portaudio.Initialize()
	if err != nil {
		panic(err.Error())
	}
	defer portaudio.Terminate()
	conn, _, err := websocket.DefaultDialer.Dial(*ip, http.Header{})
	if err != nil {
		panic(err.Error())
	}
	println("Connection opened")
	defer conn.Close()
	var stream *portaudio.Stream
	buf := new(bytes.Buffer)
	recvAudio := func(out [][]float32) {
		for i := range out[0] {
			var i16 float32
			binary.Read(buf, binary.LittleEndian, &i16)
			out[0][i] = i16
			out[1][i] = i16
		}
	}
	stream, err = portaudio.OpenDefaultStream(0, 2, 48000, 0, recvAudio)
	if err != nil {
		panic(err)
	}
	err = stream.Start()
	defer stream.Close()
	if err != nil {
		panic(err)
	}
	trm := make(chan os.Signal, 1)
	signal.Notify(trm, os.Interrupt)
	go func() {
		for {
			buf := make([]byte, 1024)
			n, _, _ := listen.ReadFrom(buf)
			buf = buf[:n]
			ttsserver.SendTTSRequest(conn, string(buf))
		}
	}()
	go func() {
		for {
			resp, err := ttsserver.RecvTTSResponse(conn)
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			binary.Write(buf, binary.LittleEndian, upscaleAudio(resp.TTSData))
		}
	}()
	keepAlive(conn, time.Second*10)
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

func upscaleAudio(in []int16) (output []float32) {
	output = make([]float32, len(in)*3)
	for o := 0; o < len(output); o += 3 {
		output[o] = float32(in[o/3]) / float32(math.MaxInt16)
		if o > 0 {
			prev := float32(in[o/3-1]) / float32(math.MaxInt16)
			cur := float32(in[o/3]) / float32(math.MaxInt16)
			delta := cur - prev
			output[o-2] = (float32(in[o/3]) / float32(math.MaxInt16)) - 2.0*(delta/3.0)
			output[o-1] = (float32(in[o/3]) / float32(math.MaxInt16)) - delta/3.0
		}
	}
	return output
}
