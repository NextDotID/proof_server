package main

import (
	"flag"
	"fmt"

	"github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/controller"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/validator/discord"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/solana"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/sirupsen/logrus"
)

var (
	flagConfigPath = flag.String("config", "./config/config.json", "Config.json file path")
	flagPort       = flag.Int("port", 9800, "Listen port")
)

func init_validators() {
	twitter.Init()
	ethereum.Init()
	keybase.Init()
	github.Init()
	discord.Init()
	solana.Init()
}

func main() {
	flag.Parse()
	config.Init(*flagConfigPath)
	logrus.SetLevel(logrus.DebugLevel)

	model.Init()
	controller.Init()
	init_validators()

	fmt.Printf("Server now running on 0.0.0.0:%d", *flagPort)
	controller.Engine.Run(fmt.Sprintf("0.0.0.0:%d", *flagPort))
}
