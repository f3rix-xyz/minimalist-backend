package main

import (
	"github.com/gin-gonic/gin"
	"github.com/youruser/yourproject/controllers"
	"github.com/youruser/yourproject/initializers"
	"github.com/youruser/yourproject/models"
)

func init() {
	// initializers.LoadEnvVariables()

	initializers.ConnectToDb()
	initializers.DB.AutoMigrate(&models.Post{})
}

func main() {

	r := gin.Default()
	r.POST("/", controllers.PostsCreate)
	r.GET(("/hello"), controllers.Hello)
	r.Run() // listen and serve on 0.0.0.0:8080
}
