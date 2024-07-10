package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/youruser/yourproject/initializers"
	"github.com/youruser/yourproject/models"
)

func PostsCreate(c *gin.Context) {

	var body struct {
		Body  string
		Title string
	}
	c.Bind(&body)
	// Code to create a post
	post := models.Post{Title: body.Title, Body: body.Body}
	result := initializers.DB.Create(&post)
	if result.Error != nil {
		c.JSON(400, gin.H{"error": result.Error})
		return
	}
	c.JSON(200, gin.H{"data": post})

}

func Hello(c *gin.Context) {
	c.JSON(200, gin.H{"data": "Hello World"})
}
