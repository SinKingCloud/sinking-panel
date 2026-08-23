package app

import "server/app/http/route"

func Run(stop <-chan struct{}) {
	route.Init(stop)
}
