package shutdown

import (
	"log"
	"sync"
	"time"
)

var waiters struct {
	sync.Mutex
	channels map[string]chan bool
	ackChan  chan string
}

func init() {
	waiters.channels = make(map[string]chan bool)
}

func Add(name string) chan bool {
	// TODO: maybe block from registering when already in shutting down state?
	waiters.Lock()
	defer waiters.Unlock()

	if ch, ok := waiters.channels[name]; ok {
		// TODO: maybe an error is more appropriate?
		return ch
	}

	log.Printf("Shutdown: registering <%s>.", name)

	waiters.channels[name] = make(chan bool, 1)
	return waiters.channels[name]
}

func Done(name string) {
	waiters.Lock()

	log.Printf("Shutdown: <%s> finished.", name)

	delete(waiters.channels, name)
	waiters.Unlock()

	if waiters.ackChan != nil {
		waiters.ackChan <- name
	}
}

func Shutdown(timeout int) []string {
	waiters.Lock()
	waiters.ackChan = make(chan string)
	waiters.Unlock()

	for name, ch := range waiters.channels {
		log.Printf("Notifying <%s>.", name)
		// send without blocking:
		select {
		case ch <- true:
		default:
		}
	}

	var timeoutCh <-chan time.Time

	if timeout > 0 {
		timeoutCh = time.After(time.Duration(timeout) * time.Millisecond)
	}

	var running []string = make([]string, 0)
	cont := true

	for cont {
		select {
		case <-waiters.ackChan:
			waiters.Lock()
			size := len(waiters.channels)
			waiters.Unlock()
			if size == 0 {
				cont = false
			}
		case <-timeoutCh:
			log.Printf("Timeout expired...")
			for name, _ := range waiters.channels {
				running = append(running, name)
			}
			cont = false
		}
	}

	return running
}
