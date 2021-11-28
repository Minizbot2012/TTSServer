package main

import (
	"bufio"
	"flag"
	"net"
	"os"
	"os/signal"
	"time"

	ttsserver "github.com/Minizbot2012/TTSServer"
	"github.com/gordonklaus/portaudio"
)

func main() {
	ip := flag.String("ip", "192.168.10.50:5555", "IP:port of the server")
	flag.Parse()
	conn, e := net.Dial("tcp", *ip)
	if e != nil {
		panic(e.Error())
	}
	conn.SetDeadline(time.Time{})
	defer conn.Close()
	println("Connection opened")
	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 1)
	stream, err := portaudio.OpenDefaultStream(0, 1, 16000, len(out), &out)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	err = stream.Start()
	brcon := bufio.NewReaderSize(conn, 65536)
	bwcon := bufio.NewWriterSize(conn, 65536)
	if err != nil {
		panic(err)
	}
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
	trm := make(chan os.Signal, 4)
	signal.Notify(trm)
	<-trm
}
