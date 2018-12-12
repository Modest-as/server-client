package main

import (
	"flag"
	"log"

	h "github.com/modest-as/server-client/src-server/handler"
	l "github.com/modest-as/server-client/src-server/listener"
	s "github.com/modest-as/server-client/src-server/store"
)

func main() {
	mode := flag.String("mode", "stateless", "server mode")
	port := flag.Int("port", 1337, "server port")

	flag.Parse()

	var handler h.Handler

	if *mode == "stateless" {
		handler = h.StatelessHandler{}
	} else if *mode == "stateful" {
		store := s.MakeInMemoryStore()
		handler = h.MakeStatefulHandler(store)
	} else {
		log.Fatalf("mode should be either stateless or stateful")
	}

	log.Println("Mode: ", *mode)
	log.Println("Port: ", *port)

	err := l.Listen(*port, handler)

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
