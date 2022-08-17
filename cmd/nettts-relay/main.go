package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
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
	devName := flag.String("out", "", "Device output name")
	list := flag.Bool("list", false, "List devices")
	flag.Parse()
	err = portaudio.Initialize()
	if err != nil {
		panic(err.Error())
	}
	defer portaudio.Terminate()
	if *list {
		dev, _ := portaudio.Devices()
		for _, deic := range dev {
			if deic.MaxOutputChannels > 0 {
				fmt.Println(deic.Name)
			}
		}
		os.Exit(0)
	}
	conn, err := net.Dial("tcp", *ip)
	if err != nil {
		panic(err.Error())
	}
	println("Connection opened")
	defer conn.Close()
	var stream *portaudio.Stream
	brcon := bufio.NewReaderSize(conn, 65536)
	bwcon := bufio.NewWriterSize(conn, 65536)
	buf := new(bytes.Buffer)
	recvAudio := func(out [][]int16) {
		for i := range out[0] {
			var i16 int16
			binary.Read(buf, binary.LittleEndian, &i16)
			out[0][i] = i16
			out[1][i] = i16
		}
	}
	if *devName == "" {
		stream, err = portaudio.OpenDefaultStream(0, 2, 48000, 0, recvAudio)
		if err != nil {
			panic(err)
		}
	} else {
		dev, _ := portaudio.Devices()
		for _, deic := range dev {
			if deic.Name == *devName {
				sp := portaudio.HighLatencyParameters(nil, deic)
				sp.Output.Channels = 2
				sp.Input.Channels = 0
				sp.SampleRate = 48000
				sp.FramesPerBuffer = 0
				stream, err = portaudio.OpenStream(sp, recvAudio)
				if err != nil {
					panic(err)
				}
				break
			}
		}
		buf.Reset()
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
			ttsserver.SendTTSRequest(bwcon, string(buf))
			bwcon.Flush()
		}
	}()
	go func() {
		for {
			resp, err := ttsserver.RecvTTSResponse(brcon)
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			binary.Write(buf, binary.LittleEndian, upscaleAudio(resp.TTSData))
		}
	}()
	<-trm
}

func upscaleAudio(in []int16) (output []int16) {
	output = make([]int16, len(in)*3)
	for i, o := 0, 0; i < len(in); i, o = i+1, o+3 {
		output[o] = in[i]
		if o > 0 && i > 0 {
			prev := in[i-1]
			cur := in[i]
			delta := int(cur - prev)
			output[o-2] = in[i] - int16(2*delta/3)
			output[o-1] = in[i] - int16(delta/3)
		}
	}
	return output
}
