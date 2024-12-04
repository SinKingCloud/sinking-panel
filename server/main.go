package main

import (
	"server/app"
	"server/bootstrap"
)

func main() {
	bootstrap.Load()
	app.Run()
}
