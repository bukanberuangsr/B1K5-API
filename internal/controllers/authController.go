package controllers

import (
	"B1K5-API/internal/utils"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type registerInput struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registeredAccount struct {
	ID         int    `json:"id"`
	CustomerID string `json:"customer_id"`
	Email      string `json:"email"`
}

func Register(ctx *gin.Context) {
	body, error := ctx.GetRawData()
	if error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
		return
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "request body is required",
		})
		return
	}

	var accounts []registerInput

	if body[0] == '[' {
		if error := json.Unmarshal(body, &accounts); error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": error.Error(),
			})
			return
		}
	} else {
		var account registerInput
		if error := json.Unmarshal(body, &account); error != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": error.Error(),
			})
			return
		}
		accounts = append(accounts, account)
	}

	if len(accounts) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one account is required",
		})
		return
	}

	registered, error := insertAccounts(accounts)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	if len(registered) == 1 {
		ctx.JSON(http.StatusOK, gin.H{
			"message":     "Register success",
			"customer_id": registered[0].CustomerID,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Register success",
		"accounts": registered,
	})
}

func insertAccounts(accounts []registerInput) ([]registeredAccount, error) {
	tx, error := utils.DB.Begin()
	if error != nil {
		return nil, error
	}
	defer tx.Rollback()

	registered := make([]registeredAccount, 0, len(accounts))

	for _, account := range accounts {
		var id int
		if error := tx.QueryRow(`SELECT nextval('customers_id_seq')`).Scan(&id); error != nil {
			return nil, error
		}

		customerID := utils.GenerateCustomerID(id)

		hash, error := utils.HashPassword(account.Password)
		if error != nil {
			return nil, error
		}

		_, error = tx.Exec(`
			INSERT INTO customers (id, customer_id, email, password_hash)
			VALUES ($1, $2, $3, $4)
		`, id, customerID, account.Email, hash)
		if error != nil {
			return nil, error
		}

		_, error = tx.Exec(`
			INSERT INTO customer_profiles (customer_id, username, full_name)
			VALUES ($1, $2, $3)
		`, id, account.Username, account.FullName)
		if error != nil {
			return nil, error
		}

		registered = append(registered, registeredAccount{
			ID:         id,
			CustomerID: customerID,
			Email:      account.Email,
		})
	}

	if error := tx.Commit(); error != nil {
		return nil, error
	}

	return registered, nil
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
		SELECT password_hash FROM customers WHERE customer_id=$1
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
		"message":     "Login success",
		"customer_id": input.CustomerID,
	})

}
