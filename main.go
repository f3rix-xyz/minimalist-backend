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
	initializers.DB.AutoMigrate(&models.User{})
}

func main() {

	r := gin.Default()
	r.GET(("/hello"), controllers.Hello)
	r.POST("/reqOTP", controllers.ReqOTP)
	r.POST("/createUser", controllers.CreateUser)
	// Protected routes
	r.POST("/login", controllers.Login)
	protected := r.Group("/")
	protected.Use(controllers.AuthMiddleware())
	{
		protected.POST("/buy", controllers.Buy)
		// other protected routes
	}
	r.Run() // listen and serve on 0.0.0.0:8080
}
