package handler

import (
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"

	pb "github.com/modest-as/server-client/grpc"
	st "github.com/modest-as/server-client/src-server/store"
)

const nUpperBound = 0xffff

// StatefulHandler stateful server handler
type StatefulHandler struct {
	store *st.Store
}

// MakeStatefulHandler returns stateful handler with store
func MakeStatefulHandler(store st.Store) StatefulHandler {
	handler := StatefulHandler{
		store: &store,
	}

	return handler
}

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
// START *uuid* *n* - starts a stream
// CONTINUE *uuid* *lastNumber*- continues existing session
func (s StatefulHandler) streamResponse(c pb.Comms_GetNumbersServer, msgAcc *messageAccessor, done *chan bool) {
	started := false
	terminate := false

	var id uuid.UUID

	for {
		if started {
			hasNext, err := (*s.store).HasNext(id)

			if err != nil {
				reply := makeErrorReply("couldn't generate next random int")
				sendReply(c, reply, done)
				return
			}

			if hasNext {
				a := (*s.store).GetNext(id)
				reply := makeDataReply(uint64(a))
				// this only guarantees that the message
				// is stored in streaming buffer
				// not that it is received by the client
				terminate = sendReply(c, reply, done)

				if !terminate {
					(*s.store).Update(id, a)
				}
			}

			if terminate || !hasNext {
				closeIfOpen(done)
				return
			}
		}

		currentMsg := msgAcc.getMsg()

		if currentMsg != "" {
			log.Println("Received: ", currentMsg)
		}

		check(currentMsg, `START (.+) (\d+)`, func(m []string) {
			if started {
				reply := makeErrorReply("server was already running")
				sendReply(c, reply, done)
				terminate = true
			}

			if len(m) != 3 {
				reply := makeErrorReply("invalid start parameters")
				sendReply(c, reply, done)
				terminate = true
			}

			var err error

			id, err = uuid.Parse(m[1])

			if err != nil {
				reply := makeErrorReply("invalid uuid")
				sendReply(c, reply, done)
				terminate = true
			}

			n, err := strconv.Atoi(m[2])

			if err != nil {
				reply := makeErrorReply("invalid n")
				sendReply(c, reply, done)
				terminate = true
			}

			if n < 1 || n > nUpperBound {
				reply := makeErrorReply("invalid n value")
				sendReply(c, reply, done)
				terminate = true
			}

			if (*s.store).Exists(id) {
				reply := makeErrorReply("uuid already taken")
				sendReply(c, reply, done)
				terminate = true
			} else {
				err = (*s.store).Add(id, n)
				if err != nil {
					reply := makeErrorReply("failed to store session")
					sendReply(c, reply, done)
					terminate = true
				}
			}

			started = true
		})

		check(currentMsg, `CONTINUE (.+) (\d+)`, func(m []string) {
			if started {
				reply := makeErrorReply("server was already running")
				sendReply(c, reply, done)
				terminate = true
			}

			if len(m) != 3 {
				reply := makeErrorReply("invalid continue parameters")
				sendReply(c, reply, done)
				terminate = true
			}

			var err error

			id, err = uuid.Parse(m[1])

			if err != nil {
				reply := makeErrorReply("invalid uuid")
				sendReply(c, reply, done)
				terminate = true
			}

			if !(*s.store).Exists(id) {
				reply := makeErrorReply("uuid has no session")
				sendReply(c, reply, done)
				terminate = true
			}

			lastNumber, err := strconv.ParseUint(m[2], 10, 32)

			if err != nil {
				reply := makeErrorReply("invalid last number")
				sendReply(c, reply, done)
				terminate = true
			}

			if !terminate {
				err = (*s.store).SyncState(id, uint32(lastNumber))

				if err != nil {
					reply := makeErrorReply("wrong last number")
					sendReply(c, reply, done)
					terminate = true
				}
			}

			if !terminate {
				count := (*s.store).GetSentCount(id)
				reply := makeDataReply(uint64(count))
				terminate = sendReply(c, reply, done)
			}

			if !terminate {
				checksum := (*s.store).GetCurrentChecksum(id)
				reply := makeDataReply(uint64(checksum))
				terminate = sendReply(c, reply, done)
			}

			started = true
		})

		if terminate {
			closeIfOpen(done)
			return
		}

		time.Sleep(period * time.Second)
	}
}
