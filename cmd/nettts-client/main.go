package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/gordonklaus/portaudio"
)

func main() {
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
		reader := bufio.NewReader(os.Stdin)
		for {
			buf, _ := reader.ReadBytes('\n')
			if buf[len(buf)-1] == '\n' {
				buf = buf[:len(buf)-1]
			}
			if buf[len(buf)-1] == '\r' {
				buf = buf[:len(buf)-1]
			}
			if string(buf) == "/exit" {
				conn.Close()
				break
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
		}
	}()
	go func() {
		for {
			buf := make([]byte, 32000)
			n, _ := conn.Read(buf)
			buf = buf[:n]
			for len(buf) > 0 {
				out[0] = int16(binary.LittleEndian.Uint16(buf[:2]))
				stream.Write()
				buf = buf[2:]
			}
		}
	}()
	trm := make(chan os.Signal, 4)
	signal.Notify(trm)
	<-trm
}
