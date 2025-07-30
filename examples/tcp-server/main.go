package main

import (
	"licklib/pkg/tcp/server"
	"log"
)

func main() {
	address := "127.0.0.1:8000"

	s, err := server.NewServer("server1", address)
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Serve(); err != nil {
		log.Println(err)
	}
	s.Close()
}
