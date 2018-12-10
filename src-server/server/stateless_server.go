package server

import (
	"log"
	"reflect"
	"sync"
	"time"

	pb "github.com/modest-as/server-client/grpc"
)

// StatelessHandler stateless server handler
type StatelessHandler struct{}

// GetNumbers handles server call to get numbers.
// We can potentially implement a state machine here.
// We also are not handling the case where GRPC context
// issues cancel signal in.
func (s StatelessHandler) GetNumbers(c pb.Comms_GetNumbersServer) error {
	var m sync.RWMutex
	var msg = ""

	done := make(chan bool)

	go listenForMsg(c, &msg, &m, &done)
	go streamResponse(c, &msg, &m, &done)

	<-done

	log.Println("Call finished!")

	return nil
}

func streamResponse(c pb.Comms_GetNumbersServer, msg *string, m *sync.RWMutex, done *chan bool) {
	started := false

	for {
		if started {
			repl := pb.Reply{Number: 2}
			err := c.Send(&repl)

			if channelIsClosed(done) {
				return
			}

			if err != nil {
				close(*done)
				logErrors(err)
			}

			log.Println("Sent: ", repl.Number)
		}

		currentMsg := getMsg(msg, m)
		if currentMsg != "" {
			log.Println("Received: ", currentMsg)
		}

		switch currentMsg {
		case "START":
			started = true
		case "END":
			close(*done)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func listenForMsg(c pb.Comms_GetNumbersServer, msg *string, m *sync.RWMutex, done *chan bool) {
	for {
		req, err := c.Recv()

		if channelIsClosed(done) {
			return
		}

		if err != nil {
			close(*done)
			logErrors(err)
		}

		m.Lock()
		*msg = req.Message
		m.Unlock()
	}
}

func getMsg(msg *string, m *sync.RWMutex) string {
	result := ""

	m.RLock()
	result = *msg
	m.RUnlock()

	if result != "" {
		m.Lock()
		*msg = ""
		m.Unlock()
	}

	return result
}

func channelIsClosed(c *chan bool) bool {
	select {
	case _ = <-*c:
		return true
	default:
		return false
	}
}

func logErrors(err error) {
	log.Fatalf("server error: %v", reflect.TypeOf(err))
}
