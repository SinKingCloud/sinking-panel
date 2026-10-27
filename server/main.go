package main

import (
	"log"
	"os"
	"server/app/command"
)

func main() {
	server := &command.Server{}
	if err := server.Execute(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
