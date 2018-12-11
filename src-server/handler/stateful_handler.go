package handler

import (
	"log"
	"strconv"
	"time"

	pb "github.com/modest-as/server-client/grpc"
)

// StatefulHandler stateful server handler
type StatefulHandler struct{}

// GetNumbers handles server call to get numbers.
// We can potentially implement a state machine here.
// We also are not handling the case where GRPC context
// issues cancel signal explicitly.
func (s StatefulHandler) GetNumbers(c pb.Comms_GetNumbersServer) error {
	msgAcc := makeMessageAccessor()

	done := make(chan bool)

	go msgAcc.listenForMsg(c, &done)
	go s.streamResponse(c, msgAcc, &done)

	<-done

	log.Println("Call finished!")

	return nil
}

// loop and send a response every *period* seconds
// Protocol:
// START - starts a stream
// CONTINUE *a* - continues from the number *a*
// END - ends the stream
func (s StatefulHandler) streamResponse(c pb.Comms_GetNumbersServer, msgAcc *messageAccessor, done *chan bool) {
	started := false
	terminate := false

	var currentVal uint64

	for {
		if started {
			reply := makeDataReply(currentVal)
			terminate = sendReply(c, reply, done)

			if terminate {
				return
			}

			currentVal *= multiplier
		}

		currentMsg := msgAcc.getMsg()

		if currentMsg != "" {
			log.Println("Received: ", currentMsg)
		}

		check(currentMsg, `START`, func(_ []string) {
			if !started {
				started = true
				currentVal = getRandomSeed()
			}
		})

		check(currentMsg, `END`, func(_ []string) {
			close(*done)
			terminate = true
		})

		check(currentMsg, `CONTINUE (\d+)`, func(m []string) {
			if started {
				reply := makeErrorReply("server was already running")
				sendReply(c, reply, done)
				terminate = true
			}

			if len(m) != 2 {
				reply := makeErrorReply("invalid continue parameters")
				sendReply(c, reply, done)
				terminate = true
			}

			val, err := strconv.ParseUint(m[1], 10, 64)

			if err != nil {
				reply := makeErrorReply("invalid continue value")
				sendReply(c, reply, done)
				terminate = true
			}

			if val == 0 {
				reply := makeErrorReply("value can't be zero")
				sendReply(c, reply, done)
				terminate = true
			}

			currentVal = val * multiplier
			started = true
		})

		if terminate {
			return
		}

		time.Sleep(period * time.Second)
	}
}
