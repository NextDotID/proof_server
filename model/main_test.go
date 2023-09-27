package model

import (
	"os"
	"testing"

	"github.com/nextdotid/proof_server/config"
)

func before_each(t *testing.T) {
	// Clean DB
	DB.Where("1 = 1").Delete(&Proof{})
	DB.Where("1 = 1").Delete(&ProofChain{})
	DB.Where("1 = 1").Delete(&AvatarAlias{})
}

func TestMain(m *testing.M) {
	config.Init("../config/config.test.json")
	Init(true)
	before_each(nil)

	os.Exit(m.Run())
}
