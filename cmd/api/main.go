package main

import (
	controller "B1K5-API/internal/controllers"
	"B1K5-API/internal/middleware"
	"B1K5-API/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitDB()
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
			* Berkaitan dengan aktivitas user seperti
			* ambil data profil, ambil data transaksi,
			* dan segment
	*/
	users := router.Group("/api/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("/", middleware.RequireRole("admin"), controller.GetAllUsers)
		users.GET("/:id", middleware.RequireSelfOrRole("admin"), controller.GetUser)
		users.GET("/:id/activity", middleware.RequireSelfOrRole("admin"), controller.GetUserActivityById)
		users.GET("/:id/segment", middleware.RequireSelfOrRole("admin"), controller.GetUserSegmentationById)
	}

	/*
		 * Rute personalisasi dan rekomendasi
			* konfigurasi pengguna terkait personalisasi
			* konten/fitur dan rekomendasi konten/fitur
			* ke pengguna
	*/
	personalization := router.Group("/api/personalization")
	personalization.Use(middleware.AuthMiddleware())
	{
		personalization.GET("/:id", middleware.RequireSelfOrRole("admin"), controller.GetPersonalizationById)
	}

	recommendations := router.Group("/api/recommendations")
	recommendations.Use(middleware.AuthMiddleware())
	{
		recommendations.GET("/:id", middleware.RequireSelfOrRole("admin"), controller.GetRecommendationByUserID)
	}

	recommendation := router.Group("/api/recommendation")
	recommendation.Use(middleware.AuthMiddleware())
	{
		recommendation.GET("/:id", middleware.RequireSelfOrRole("admin"), controller.GetRecommendationByUserID)
	}

	/*
		 * Rute analytics
			* menyimpan analytics semua aktifitas dalam
			* sebuah file/database (misalnya mongodb)
	*/
	analytics := router.Group("/api/analytics")
	analytics.Use(middleware.AuthMiddleware())
	{
		analytics.GET("/metrics", middleware.RequireRole("admin"), controller.GetAnalyticsMetrics)
		analytics.POST("/event", controller.CreateAnalyticsEvent)
	}

	router.POST(
		"/api/segments/update",
		middleware.AuthMiddleware(),
		middleware.RequireRole("admin"),
		controller.UpdateUserSegments,
	)

	router.Run(":8080")
}
