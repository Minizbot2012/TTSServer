package main

import (
	"flag"
	"net"
	"os"
	"os/signal"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/gordonklaus/portaudio"
)

func main() {
	listen, err := net.ListenPacket("udp", ":5555")
	if err != nil {
		panic(err)
	}
	ip := flag.String("ip", "192.168.10.50:5555", "IP:port of the server")
	flag.Parse()
	conn, e := net.Dial("tcp", *ip)
	if e != nil {
		panic(e.Error())
	}
	println("Connection opened")
	defer conn.Close()
	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 1)
	stream, err := portaudio.OpenDefaultStream(0, 1, 16000, len(out), &out)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	err = stream.Start()
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			buf := make([]byte, 1024)
			n, _, _ := listen.ReadFrom(buf)
			buf = buf[:n]
			conn.Write(buf)
		}
	}()
	go func() {
		for {
			resp, err := ttsserver.RecvTTSResponse(conn)
			if err != nil {
				println(e.Error())
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
	trm := make(chan os.Signal, 4)
	signal.Notify(trm)
	<-trm
}
