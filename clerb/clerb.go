package clerb

import (
	"log"
	"os/exec"
)

type ClerbWindow struct {
	W *exec.Cmd
	E chan bool
}

func (cw *ClerbWindow) Wait() {
	cw.W.Wait()
	log.Printf("Main process exited, sending quit signal")
	cw.E <- true
}

func StartClerbWindow() ClerbWindow {
	x := ClerbWindow{
		exec.Command("nw", ".", "--remote-debugging-port=7374"),
		make(chan bool, 0),
	}

	x.W.Run()
	go x.Wait()
	return x
}

func Run() {
	server := NewServer(":7373")
	server.BindAll()
	go server.Serve()

	w := StartClerbWindow()
	<-w.E
}
