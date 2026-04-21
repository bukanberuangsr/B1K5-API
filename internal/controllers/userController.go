package controllers

import (
	"B1K5-API/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func activity(ctx *gin.Context) {
	var user models.Customer
	ctx.JSON(http.StatusOK, gin.H{
		"message": "user controller is running!",
		"user":    user,
	})
}
