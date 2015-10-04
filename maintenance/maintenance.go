package maintenance

import (
	"log"
	"sync"
	"time"
)

var ticker *time.Ticker
var stopper chan bool

var handlers = []MaintenaceHandler{}
var handlerslock sync.Mutex

// Start will start up the regualr ticking of maintenance handlers.
func Start(interval, delay time.Duration) {
	ticker = time.NewTicker(interval)
	stopper = make(chan bool)

	go func() {
		<-time.After(delay)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					log.Println("Maintenace tick")
					handleTick()
				case <-stopper:
					return
				}
			}
		}()
	}()
	log.Println("Maintenace is started up.")
}

func handleTick() {
	for _, handler := range handlers {
		log.Println("Handler started")
		err := handler.Execute()
		if err != nil {
			log.Println(err)
		}
		//if handler.RemoveAfterExecute() {
		//	handler = append(handler[:i], handler[i+1:]...)
		//}
	}
}

// Register will add a MaintenaceHandler in the maintenance queue.
func Register(handler MaintenaceHandler) {
	handlerslock.Lock()
	defer handlerslock.Unlock()
	handlers = append(handlers, handler)
}

// Stop function will stop the regular ticking of maintenance handlers.
func Stop() {
	stopper <- true
}
