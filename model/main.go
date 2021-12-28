package model

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nextdotid/proof-server/config"
)

var (
	DB *gorm.DB
	l  = logrus.WithFields(logrus.Fields{"module": "model"})
)

// Init initializes DB connection instance and do migration at startup.
func Init() {
	if DB != nil { // initialized
		return
	}
	dsn := config.GetDatabaseDSN()
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Fatalf("Error when opening DB: %s\n", err.Error())
	}

	err = DB.AutoMigrate(
		&Proof{},
		&ProofChain{},
	)
	if err != nil {
		panic(err)
	}

	l.Info("database initialized")
}
