package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	config_db "github.com/jay6909/dino-internal-wallet-service/internal/config/db"
	config_env "github.com/jay6909/dino-internal-wallet-service/internal/config/env"
	"github.com/jay6909/dino-internal-wallet-service/internal/data/repository"
	"github.com/jay6909/dino-internal-wallet-service/internal/handler"
	"github.com/jay6909/dino-internal-wallet-service/internal/seed"
)

var appEnv *config_env.AppEnv
var db config_db.DB

func main() {
	var err error
	appEnv, err = config_env.LoadAppEnv()
	if err != nil {
		panic(err)
	}

	db, err = config_db.NewGormDB(appEnv.DatabaseConfig.DSN)
	if err != nil {
		panic(err)
	}

	//migrate db
	db.Migrate()

	if appEnv.Seed {
		seed.SeedDb()
		fmt.Println("✅ Seeding complete")

		return
	}

	r := gin.Default()

	//init repositories
	userRepository := repository.NewUserRepository(db.GetDB())
	walletRepository := repository.NewWalletRepository(db.GetDB())

	//init handlers
	userHandler := handler.NewUserHandler(userRepository)
	walletHandler := handler.NewWalletHandler(walletRepository, userRepository)
	apiV1 := r.Group("/api/v1")
	{
		userHandler.RegisterRoutes(apiV1)
		walletHandler.RegisterRoutes(apiV1)
	}

	r.Run(fmt.Sprintf(":%s", appEnv.Port))

}
