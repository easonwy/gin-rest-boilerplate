package main

import (
	"fmt"
	"log"

	"github.com/example/go-user-service/internal/user/model"
)

func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Auto-migrate the User model
	err = app.DB.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	port := app.Config.App.Port
	log.Printf("Starting server on :%d", port)
	app.Router.Run(fmt.Sprintf(":%d", port))
}
