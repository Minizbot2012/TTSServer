package main

import (
	"encoding/binary"
	"os"

	"github.com/gordonklaus/portaudio"
)

func main() {
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
	for {
		binary.Read(os.Stdin, binary.LittleEndian, &out)
		stream.Write()
	}
}
