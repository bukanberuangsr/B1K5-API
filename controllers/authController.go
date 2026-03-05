package controller

import (
	"B1K5-API/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register dummy
func Register(ctx *gin.Context) {
	var input models.User

	if error := ctx.ShouldBindJSON(&input); error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
		return
	}
	input.ID = len(models.Users) + 1
	models.Users = append(models.Users, input)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Register success",
		"user":    input,
	})
}

// Login dummy
func Login(ctx *gin.Context) {
	var input models.User

	if error := ctx.ShouldBindJSON(&input); error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
	}

	for _, user := range models.Users {
		if user.Email == input.Email && user.Password == input.Password {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "Login success",
				"user":    user,
			})
		}
	}

}
