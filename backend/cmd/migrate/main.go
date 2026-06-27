package main

import (
	"log"

	"bitbucket.org/HyperEther/document-management-system/internal/config"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/db"
)

func main() {
	cfg := config.Load()
	database, _ := db.NewPostgresConnection(cfg.DB)
	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}
	log.Print("Success")
}
