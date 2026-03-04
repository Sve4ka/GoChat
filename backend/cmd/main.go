package main

import (
	"backend/internal/delivery"
	"backend/pkg/config"
	"backend/pkg/log"
	"backend/pkg/postgres"
)

func main() {

	cfg := config.InitConfig()

	log.Log.Info("Config Initialized")

	db := postgres.MustInitPg(cfg)
	defer db.Close()

	log.Log.Info("PG Initialized")

	delivery.Start(db)

}
