package main

import (
	"log"
	"os"

	"server/app/command"
)

func main() {
	server, err := command.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	if err = server.Execute(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
