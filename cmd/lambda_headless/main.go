package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/nextdotid/proof_server/headless"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	headless.Init("/opt/chromium")
}

func main() {
	algnhsa.ListenAndServe(headless.Engine, nil)
}
