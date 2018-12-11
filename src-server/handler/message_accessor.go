package handler

import (
	"fmt"
	"regexp"
	"sync"

	pb "github.com/modest-as/server-client/grpc"
)

type messageAccessor struct {
	msg  *string
	lock *sync.RWMutex
}

func makeMessageAccessor() *messageAccessor {
	var msg string
	var lock sync.RWMutex

	msgAcc := messageAccessor{
		msg:  &msg,
		lock: &lock,
	}

	return &msgAcc
}

// listen for all messages from the client
// note that if the server didn't act on the
// operation that is already in memory and
// a new one comes in, that message will
// override the previous one. I.e. message
// execution is not guaranteed
func (accessor *messageAccessor) listenForMsg(c pb.Comms_GetNumbersServer, done *chan bool) {
	for {
		req, err := c.Recv()

		if channelIsClosed(done) {
			return
		}

		if err != nil {
			close(*done)
			logErrors(err)
			return
		}

		accessor.lock.Lock()
		*accessor.msg = req.Message
		accessor.lock.Unlock()
	}
}

// safely check the last message from the client
func (accessor *messageAccessor) getMsg() string {
	result := ""

	accessor.lock.RLock()
	result = *accessor.msg
	accessor.lock.RUnlock()

	if result != "" {
		accessor.lock.Lock()
		*accessor.msg = ""
		accessor.lock.Unlock()
	}

	return result
}

// validate message pattern and execute an action if it matches
func check(message string, pattern string, action func([]string)) {
	r := regexp.MustCompile(fmt.Sprintf("^%s$", pattern))

	matches := r.FindStringSubmatch(message)

	if len(matches) == 0 {
		return
	}

	action(matches)
}
