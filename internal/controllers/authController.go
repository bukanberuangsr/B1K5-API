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
	Role       string `json:"role"`
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
			"role":        registered[0].Role,
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
		role := "customer"

		hash, error := utils.HashPassword(account.Password)
		if error != nil {
			return nil, error
		}

		_, error = tx.Exec(`
			INSERT INTO customers (id, customer_id, username, email, password_hash, role)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, id, customerID, account.Username, account.Email, hash, role)
		if error != nil {
			return nil, error
		}

		_, error = tx.Exec(`
			INSERT INTO customer_profiles (customer_id, full_name)
			VALUES ($1, $2)
		`, id, account.FullName)
		if error != nil {
			return nil, error
		}

		registered = append(registered, registeredAccount{
			ID:         id,
			CustomerID: customerID,
			Email:      account.Email,
			Role:       role,
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

	var userID int
	var customerID string
	var hash string
	var role string

	error := utils.DB.QueryRow(`
		SELECT id, customer_id, password_hash, role
		FROM customers
		WHERE customer_id = $1
	`, input.CustomerID).Scan(&userID, &customerID, &hash, &role)

	if error != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid customer_id or password",
		})
		return
	}

	if !utils.CheckPassword(hash, input.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong password",
		})
		return
	}

	token, error := utils.GenerateJWT(userID, customerID, role)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Login success",
		"customer_id": customerID,
		"role":        role,
		"token":       token,
	})

}

func LoginById(ctx *gin.Context) {
	var input struct {
		CustomerID string `json:"customer_id"`
	}

	if error := ctx.ShouldBindJSON(&input); error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": error.Error(),
		})
		return
	}

	var userID int
	var customerID string
	// var hash string
	var role string

	error := utils.DB.QueryRow(`
		SELECT id, customer_id, role
		FROM customers
		WHERE customer_id = $1
	`, input.CustomerID).Scan(&userID, &customerID, &role)

	if error != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid customer_id or password",
		})
		return
	}

	// if !utils.CheckPassword(hash, input.Password) {
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{
	// 		"error": "Wrong password",
	// 	})
	// 	return
	// }

	token, error := utils.GenerateJWT(userID, customerID, role)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Login success",
		"customer_id": customerID,
		"role":        role,
		"token":       token,
	})

}
