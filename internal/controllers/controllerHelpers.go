package controllers

import (
	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

type customerIdentity struct {
	ID         int
	CustomerID string
}

func getCustomerByIDParam(id string) (customerIdentity, error) {
	var customer customerIdentity

	err := utils.DB.QueryRow(`
		SELECT id, customer_id
		FROM customers
		WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&customer.ID, &customer.CustomerID)

	return customer, err
}

type userActivityData struct {
	CustomerID     string  `json:"customer_id"`
	Transactions   []gin.H `json:"transactions"`
	FrequentlyUsed []gin.H `json:"frequently_used_features"`
}

func getUserActivityData(id string) (userActivityData, error) {
	var data userActivityData

	if err := utils.DB.QueryRow(`
		SELECT customer_id FROM customers WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&data.CustomerID); err != nil {
		return data, err
	}

	transactionRows, err := utils.DB.Query(`
		SELECT
			t.trx_id,
			t.type,
			t.amount,
			t.status,
			t.created_at::text
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN customers c ON c.id = a.customer_id
		WHERE c.id::text = $1 OR c.customer_id = $1
		ORDER BY t.created_at DESC
	`, id)
	if err != nil {
		return data, err
	}
	defer transactionRows.Close()

	data.Transactions = []gin.H{}

	for transactionRows.Next() {
		var trxID, trxType, status, createdAt string
		var amount float64

		if err := transactionRows.Scan(&trxID, &trxType, &amount, &status, &createdAt); err != nil {
			return data, err
		}

		data.Transactions = append(data.Transactions, gin.H{
			"trx_id":     trxID,
			"type":       trxType,
			"amount":     amount,
			"status":     status,
			"created_at": createdAt,
		})
	}

	if err := transactionRows.Err(); err != nil {
		return data, err
	}

	featureRows, err := utils.DB.Query(`
		SELECT
		    ua.feature,
		    COUNT(*) AS usage_count,
		    MAX(ua.created_at)::text AS last_used_at
		FROM user_activities ua
		JOIN customers c ON c.id = ua.customer_id
		WHERE c.id::text = $1 OR c.customer_id = $1
		GROUP BY ua.feature
		ORDER BY usage_count DESC, MAX(ua.created_at) DESC
	`, id)
	if err != nil {
		return data, err
	}
	defer featureRows.Close()

	data.FrequentlyUsed = []gin.H{}

	for featureRows.Next() {
		var feature, lastUsedAt string
		var usageCount int

		if err := featureRows.Scan(&feature, &usageCount, &lastUsedAt); err != nil {
			return data, err
		}

		data.FrequentlyUsed = append(data.FrequentlyUsed, gin.H{
			"feature":      feature,
			"usage_count":  usageCount,
			"last_used_at": lastUsedAt,
		})
	}

	if err := featureRows.Err(); err != nil {
		return data, err
	}

	return data, nil
}
