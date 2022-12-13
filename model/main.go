package model

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nextdotid/proof_server/config"
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
	dsn := config.GetDatabaseDSN(config.C.DB.Host)
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

func GetReadOnlyDB() *gorm.DB {
	host := lo.Sample(config.C.DB.ReadOnlyHosts)
	dsn := config.GetDatabaseDSN(host)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
