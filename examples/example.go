package main

import (
	"github.com/arikfr/go-shutdown/shutdown"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ch := shutdown.Add("test")
	ch2 := shutdown.Add("test2")

	go func() {
		for {
			<-ch2
			time.Sleep(time.Duration(800) * time.Millisecond)
			log.Printf("Received the signal")
			shutdown.Done("test2")
		}
	}()

	go func() {
		for {
			<-ch
			log.Printf("Received the signal")
			shutdown.Done("test")
		}
	}()

	waitForSignal()

	running := shutdown.Shutdown(1000)
	if len(running) == 0 {
		log.Printf("All stopped")
	} else {
		log.Printf("Still running: %s", running)
	}

}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)
	<-c
}
