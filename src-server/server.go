package main

import (
	"flag"
	"log"

	sv "github.com/modest-as/server-client/src-server/handler"
	ls "github.com/modest-as/server-client/src-server/listener"
)

func main() {
	mode := flag.String("mode", "stateless", "server mode")
	port := flag.Int("port", 1337, "server port")

	flag.Parse()

	var handler sv.Handler

	if *mode == "stateless" {
		handler = sv.StatelessHandler{}
	} else if *mode == "stateful" {
		handler = sv.StatefulHandler{}
	} else {
		log.Fatalf("mode should be either stateless or stateful")
	}

	log.Println("Mode: ", *mode)
	log.Println("Port: ", *port)

	err := ls.Listen(*port, handler)

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
