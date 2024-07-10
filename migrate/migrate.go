package main

import (
	"github.com/youruser/yourproject/initializers"
	"github.com/youruser/yourproject/models"
)

func init() {
	// Connect to the database
	// initializers.LoadEnvVariables()
	initializers.ConnectToDb()

}

func main() {
	// Run the migrations
	initializers.DB.AutoMigrate(&models.Post{})
}
