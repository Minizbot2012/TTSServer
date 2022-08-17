package main

import (
	"bufio"
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
	out := make([]int16, 48000/5)
	var stream *portaudio.Stream
	if *devName == "" {
		stream, err = portaudio.OpenDefaultStream(0, 1, 48000, len(out), &out)
		if err != nil {
			panic(err)
		}
	} else {
		dev, _ := portaudio.Devices()
		for _, deic := range dev {
			if deic.Name == *devName {
				sp := portaudio.HighLatencyParameters(nil, deic)
				sp.Output.Channels = 1
				sp.Input.Channels = 0
				sp.SampleRate = 48000
				sp.FramesPerBuffer = len(out)
				stream, err = portaudio.OpenStream(sp, &out)
				if err != nil {
					panic(err)
				}
				break
			}
		}
	}
	err = stream.Start()
	defer stream.Close()
	if err != nil {
		panic(err)
	}
	trm := make(chan os.Signal, 48000/5)
	signal.Notify(trm, os.Interrupt)
	brcon := bufio.NewReaderSize(conn, 65536)
	bwcon := bufio.NewWriterSize(conn, 65536)
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
		bw := bufwriter{out, stream, 0}
		if err != nil {
			panic(err)
		}
		for {
			resp, err := ttsserver.RecvTTSResponse(brcon)
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			bw.processSpeechData(resp.TTSData)
		}
	}()
	<-trm
}

type bufwriter struct {
	output []int16
	stream *portaudio.Stream
	pos    int
}

func (b *bufwriter) processSpeechData(input []int16) {
	rem := len(input)
	offset := 0
	for rem > 0 {
		// copy our input speech data to the portaudio buffer
		// upsample from 16000 hz to 48000 hz
		n := 0
		in, out := offset, b.pos
		for {
			if in >= len(input) {
				break
			}
			if out >= len(b.output) {
				break
			}
			b.output[out] = input[in]
			if out > 0 && in > 0 {
				prev := input[in-1]
				cur := input[in]
				delta := int(cur - prev)
				b.output[out-2] = input[in] - int16(2*delta/3)
				b.output[out-1] = input[in] - int16(delta/3)
			}
			in += 1
			out += 3
			n++
		}
		b.pos = out
		rem -= n
		offset += n
		if n == 0 {
			if err := b.stream.Write(); err != nil {
				fmt.Println(err)
			}
			b.pos = 0
			// portaudio has the data now, so clear our buffer to prevent a
			// flush in main from replaying remaining data.
			for i := 0; i < len(b.output); i++ {
				b.output[i] = 0
			}
		}
	}
}
