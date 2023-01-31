package main

import (
	"flag"
	"fmt"

	"github.com/nextdotid/proof_server/headless"
	"github.com/sirupsen/logrus"
)

var (
	flagPort = flag.Int("port", 9801, "Listen port")
)

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	headless.Init("")

	fmt.Printf("Server now running on 0.0.0.0:%d", *flagPort)
	headless.Engine.Run(fmt.Sprintf("0.0.0.0:%d", *flagPort))
}
