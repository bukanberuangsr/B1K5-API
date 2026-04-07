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

	/*
		 * Rute otentikasi
			* Autentikasi pengguna disarankan menggunakan JWTAUTH.
		 	* Untuk enpointnya nunggu PM (Vina)
			* Sementara cocokan dengan endpoint yang ada dibuat
	*/
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", controller.Register)
		auth.POST("/login", controller.Login)
	}

	/*
		 * Rute user
			* Berkaitan dengan aktifitas user seperti
			* ambil data profil, ambil data transaksi,
			* dan segment
	*/
	users := router.Group("/api/users")
	{
		users.GET("/:id")
		users.GET("/:id/activity")
		users.GET("/:id/segment")
	}

	/*
		 * Rute personalisasi dan rekomendasi
			* konfigurasi pengguna terkait personalisasi
			* konten/fitur dan rekomendasi konten/fitur
			* ke pengguna
	*/
	personalization := router.Group("/api/personalization")
	{
		personalization.GET("/:id")
	}

	recommendation := router.Group("/api/recommendation")
	{
		recommendation.GET("/:id")
	}

	/*
		 * Rute analytics
			* menyimpan analytics semua aktifitas dalam
			* sebuah file/database (misalnya mongodb)
	*/
	analytics := router.Group("/api/analytics")
	{
		analytics.GET("/metrics") // TODO: protect with auth
		analytics.POST("/event")
	}

	router.POST("/api/segments/update", func(ctx *gin.Context) {
		// TODO: update segmentasi dari divisi AI/ML
	})

	router.Run(":8080")
}
