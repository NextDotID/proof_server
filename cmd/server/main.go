package main

import (
	"flag"
	"fmt"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/model"
	"github.com/sirupsen/logrus"
)

var flagConfigPath = flag.String("config", "./config/config.json", "Config.json file path")
var flagPort = flag.Int("port", 9800, "Listen port")

func main() {
	flag.Parse()
	config.Init(*flagConfigPath)
	logrus.SetLevel(logrus.DebugLevel)

	model.Init()
	controller.Init()

	fmt.Printf("Server now running on 0.0.0.0:%d", *flagPort)
	controller.Engine.Run(fmt.Sprintf("0.0.0.0:%d", *flagPort))
}
