package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/nextdotid/proof_server/common"
	"github.com/nextdotid/proof_server/headless"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	common.CurrentRuntime = common.Runtimes.Lambda
	headless.Init("/opt/chromium", "")
}

func main() {
	algnhsa.ListenAndServe(headless.Engine, nil)
}
