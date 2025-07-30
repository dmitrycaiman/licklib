package main

import (
	"licklib/pkg/udp/server"
	"log"
)

func main() {
	address := "127.0.0.1:12345"

	s, err := server.NewServer("server1", address)
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Serve(); err != nil {
		log.Println(err)
	}
}
