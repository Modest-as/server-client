package main

import (
	"flag"
	"log"

	sv "github.com/modest-as/server-client/src-server/server"
)

func main() {
	mode := flag.String("mode", "stateless", "server mode")
	port := flag.Int("port", 1337, "server port")

	flag.Parse()

	var handler sv.Handler

	if *mode == "stateless" {
		handler = sv.StatelessHandler{}
	}

	log.Println("Mode: ", *mode)
	log.Println("Port: ", *port)

	err := Listen(*port, handler)

	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
