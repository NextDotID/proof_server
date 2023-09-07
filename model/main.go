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
	// Since this service is mostly run as a lambda, we don't need
	// to init an array here.  When lambda scaled to a very large
	// number, servers in `read_only_hosts` will be used evenly.
	ReadOnlyDB *gorm.DB
	l          = logrus.WithFields(logrus.Fields{"module": "model"})
)

// Init initializes DB connection instance and do migration at startup.
func Init(autoMigrate bool) {
	if DB != nil { // initialized
		return
	}
	dsn := config.GetDatabaseDSN(config.C.DB.Host)
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		l.Fatalf("Error when opening DB: %s\n", err.Error())
	}

	if autoMigrate {
		err = DB.AutoMigrate(
			&Proof{},
			&ProofChain{},
		)
		if err != nil {
			panic(err)
		}
	}

	readOnlyHost := lo.Sample(config.C.DB.ReadOnlyHosts)
	readOnlyDSN := config.GetDatabaseDSN(readOnlyHost)
	ReadOnlyDB, err = gorm.Open(postgres.Open(readOnlyDSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	l.Info("database initialized")
}
