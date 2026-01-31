package main

import (
	"log"

	"github.com/emanuele-dedonatis/lorawan-simulator/internal/cli"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/networkserver"
)

func main() {
	pool := networkserver.NewPool()

	if err := cli.InitRootCmd(pool).Execute(); err != nil {
		log.Fatal(err)
	}
}
