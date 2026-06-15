package main

import (
	"let-it-spin/internal/config"
	"let-it-spin/internal/database"
	"let-it-spin/router"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := router.SetupRouter(db)
	r.Run(":" + cfg.AppPort)
}
