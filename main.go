package main

import (
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
	// login JWT token (hiks)
	router.GET("/api/auth/login", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "login",
		})
	})

	// register
	router.GET("/api/auth/register", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "register",
		})
	})

	// Rute User

	// Rute Consent

	// Rute Personalisasi

	// Rute Admin

	router.Run() // listens on port :8080
}
