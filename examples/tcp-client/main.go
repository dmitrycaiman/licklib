package main

import (
	"bufio"
	"fmt"
	"licklib/pkg/tcp/client"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	address := "127.0.0.1:8000"

	c, err := client.NewClient("client1", address)
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Connect(50); err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := c.Read(); err != nil {
			log.Println(err)
		}
	}()

	scn := bufio.NewScanner(os.Stdin)
	for {
		if !scn.Scan() {
			log.Fatal(scn.Err())
		}

		t := scn.Text()
		switch {
		case strings.HasPrefix(t, "exit"):
			c.Close()
			return
		case strings.HasPrefix(t, "batch "):
			n, err := strconv.Atoi(strings.TrimLeft(t, "batch "))
			if err != nil {
				t = fmt.Sprintf("failed to read batch size: %v", err)
			} else {
				t = string(make([]byte, n))
			}
		}

		if err := c.Send(t); err != nil {
			log.Println(err)
		}
	}
}
