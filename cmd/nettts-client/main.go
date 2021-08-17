package main

import (
	"bufio"
	"flag"
	"net"
	"os"
	"os/exec"
	"os/signal"

	ttsserver "github.com/Minizbot2012/TTSServer"
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
	cmdAPlay := exec.Command("ttsplay.exe")
	r := ttsserver.NewNotify(conn)
	cmdAPlay.Stdin = r
	e = cmdAPlay.Start()
	if e != nil {
		panic(e.Error())
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
				r.C <- true
				break
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
		}
	}()
	trm := make(chan os.Signal, 4)
	signal.Notify(trm)
	select {
	case <-r.C:
		println("Socket closed, shutting down")
		break
	case <-trm:
		break
	}
	cmdAPlay.Process.Kill()
	cmdAPlay.Wait()
}
