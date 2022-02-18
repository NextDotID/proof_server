package model

import (
	"os"
	"testing"

	"github.com/nextdotid/proof-server/config"
)

func before_each(t *testing.T) {
	// Clean DB
	DB.Where("1 = 1").Delete(&Proof{})
	DB.Where("1 = 1").Delete(&ProofChain{})
	DB.Where("1 = 1").Delete(&KV{})
}

func TestMain(m *testing.M) {
	config.Init("../config/config.test.json")
	Init()
	before_each(nil)

	os.Exit(m.Run())
}
