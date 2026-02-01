package main

import (
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/api"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
)

func main() {
	pool := networkserver.NewPool()

	api.Init(pool)
}
