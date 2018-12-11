package handler

import (
	"log"
	"strconv"

	pb "github.com/modest-as/server-client/grpc"
)

// check if streaming should stop
func channelIsClosed(c *chan bool) bool {
	select {
	case _ = <-*c:
		return true
	default:
		return false
	}
}

func makeDataReply(number uint64) *pb.Reply {
	data := pb.Data{Number: number}
	replyData := pb.Reply_Data{Data: &data}
	return &pb.Reply{Payload: &replyData}
}

func makeErrorReply(message string) *pb.Reply {
	error := pb.Error{Message: message}
	replyError := pb.Reply_Error{Error: &error}
	return &pb.Reply{Payload: &replyError}
}

func logErrors(err error) {
	log.Printf("server error: %v", err)
}

func logReply(reply *pb.Reply) {
	err := reply.GetError()

	message := ""

	if err != nil {
		message = err.Message
	} else {
		message = strconv.FormatUint(reply.GetData().Number, 10)
	}

	log.Println("Sent: ", message)
}
