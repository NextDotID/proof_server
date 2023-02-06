package main

import (
	"flag"
	"fmt"

	"github.com/nextdotid/proof_server/common"
	"github.com/nextdotid/proof_server/headless"
	"github.com/sirupsen/logrus"
)

var (
	flagPort         = flag.Int("port", 9801, "Listen port")
	flagChromiumPath = flag.String("chromium", "/usr/bin/chromium-browser", "Path to Chromium executable")
)

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	common.CurrentRuntime = common.Runtimes.Standalone
	headless.Init(*flagChromiumPath)

	listen := fmt.Sprintf("0.0.0.0:%d", *flagPort)
	headless.Engine.Run(listen)
}
