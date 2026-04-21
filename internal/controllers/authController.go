package controllers

import (
	"B1K5-API/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(ctx *gin.Context) {
	var input struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if error := ctx.ShouldBindJSON(&input); error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
		return
	}

	hash, error := utils.HashPassword(input.Password)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	// TODO: buat agar customerID (string) di generate dan dimasukkan tabel customers
	var id int

	error = utils.DB.QueryRow(`
		INSERT INTO customers (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`, input.Email, hash).Scan(&id)

	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	// Generate CustomerID (string) untuk login
	customerID := utils.GenerateCustomerID(id)

	_, error = utils.DB.Exec(`
		INSERT INTO customer_profiles (customer_id, full_name)
		VALUES ($1, $2)
	`, id, input.FullName)

	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Register success",
		"customer_id": customerID,
	})
}

func Login(ctx *gin.Context) {
	var input struct {
		CustomerID string `json:"customer_id"`
		Password   string `json:"password"`
	}

	if error := ctx.ShouldBindJSON(&input); error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
		return
	}

	var hash string

	error := utils.DB.QueryRow(`
		SELECT pasword_hash FROM customers WHERE customer_id=$1
	`, input.CustomerID).Scan(&hash)

	if error != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": error.Error(),
		})
		return
	}

	if !utils.CheckPassword(hash, input.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong password",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login success",
	})

}
