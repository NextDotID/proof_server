package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/sirupsen/logrus"
	"github.com/nextdotid/proof_server/headless"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	headless.Init("/opt/chromium")
}

func main() {
	algnhsa.ListenAndServe(headless.Engine, nil)
}

