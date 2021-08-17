package main

import (
	"net"
	"os"
	"os/exec"
	"time"

	ttsserver "github.com/Minizbot2012/TTSServer"
)

func main() {
	listen, _ := net.Listen("tcp", "0.0.0.0:5555")
	for {
		conn, err := listen.Accept()
		if err != nil {
			println(err.Error())
		} else {
			go handleConn(conn)
		}
	}
}

func handleConn(conn net.Conn) {
	println("New Connection")
	cmd := exec.Command("pico-tts")
	r := ttsserver.NewNotify(conn)
	cmd.Stdin = r
	cmd.Stdout = conn
	go func() {
		err := cmd.Run()
		if err != nil {
			println(err.Error())
		}
	}()
	tkr := time.NewTicker(time.Second * 30)
loop:
	for {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			r.C <- true
			break loop
		}
		select {
		case <-tkr.C:
			continue loop
		case <-r.C:
			break loop
		}
	}
	tkr.Stop()
	close(r.C)
	conn.Close()
	println("Connection Closed")
	cmd.Process.Signal(os.Interrupt)
	cmd.Process.Wait()
}
