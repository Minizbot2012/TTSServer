package main

import (
	"net"

	"github.com/gordonklaus/portaudio"
	"github.com/tzneal/gopicotts"
)

func main() {
	listen, err := net.ListenPacket("udp", ":5555")
	if err != nil {
		panic(err)
	}
	ttsEngine, _ := gopicotts.NewEngine(gopicotts.DefaultOptions)
	defer ttsEngine.Close()
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
	ttsEngine.SetOutput(func(c []int16) {
		for _, v := range c {
			out[0] = v
			stream.Write()
		}
	})
	for {
		buf := make([]byte, 1024)
		n, _, _ := listen.ReadFrom(buf)
		ttsEngine.SendText(string(buf[:n]))
		ttsEngine.FlushSendText()
	}
}
