package main

import (
	controller "B1K5-API/internal/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/api/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "The API is currently running!",
		})
	})

	// Rute otentikasi

	auth := router.Group("api/auth")
	{
		auth.POST("/register", controller.Register)
		auth.POST("/login", controller.Login)
	}

	router.Run(":8000")
}
