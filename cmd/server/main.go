package main

import (
	"flag"
	"fmt"

	"github.com/nextdotid/proof_server/common"
	"github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/controller"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/validator/activitypub"
	"github.com/nextdotid/proof_server/validator/das"
	"github.com/nextdotid/proof_server/validator/discord"
	"github.com/nextdotid/proof_server/validator/dns"
	"github.com/nextdotid/proof_server/validator/ethereum"
	"github.com/nextdotid/proof_server/validator/github"
	"github.com/nextdotid/proof_server/validator/keybase"
	"github.com/nextdotid/proof_server/validator/minds"
	"github.com/nextdotid/proof_server/validator/solana"
	"github.com/nextdotid/proof_server/validator/steam"
	"github.com/nextdotid/proof_server/validator/twitter"
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
	das.Init()
	solana.Init()
	minds.Init()
	dns.Init()
	steam.Init()
	activitypub.Init()
}

func main() {
	flag.Parse()
	config.Init(*flagConfigPath)
	logrus.SetLevel(logrus.DebugLevel)
	common.CurrentRuntime = common.Runtimes.Standalone

	model.Init()
	controller.Init()
	init_validators()

	fmt.Printf("Server now running on 0.0.0.0:%d", *flagPort)
	controller.Engine.Run(fmt.Sprintf("0.0.0.0:%d", *flagPort))
}
