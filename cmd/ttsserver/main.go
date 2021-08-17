package main

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
)

type reader struct {
	conn net.PacketConn
}

func (r *reader) Read(d []byte) (int, error) {
	b, _, e := r.conn.ReadFrom(d)
	return b, e
}

func main() {
	listen, err := net.ListenPacket("udp", ":5555")
	if err != nil {
		panic(err)
	}

	cmdTTS := exec.Command("pico-tts")
	cmdTTS.Stdin = &reader{conn: listen}
	ttso, _ := cmdTTS.StdoutPipe()
	go func() {
		err := cmdTTS.Run()
		panic(err)
	}()
	if err != nil {
		panic(err)
	}
	cmdAPlay := exec.Command("ttsplay")
	cmdAPlay.Stdin = ttso
	cmdAPlay.Stderr = os.Stdout
	cmdAPlay.Stdout = os.Stdout
	go func() {
		err := cmdAPlay.Run()
		panic(err)
	}()
	if err != nil {
		panic(err)
	}
	var kill = make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)
	<-kill
	if err != nil {
		panic(err)
	}
	cmdAPlay.Process.Kill()
	cmdTTS.Process.Kill()
}
