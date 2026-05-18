package controllers

import (
	"B1K5-API/internal/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userResponse struct {
	ID         int    `json:"id"`
	CustomerID string `json:"customer_id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	CreatedAt  string `json:"created_at"`
}

type segmentResponse struct {
	ID              int                      `json:"id"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	Confidence      float64                  `json:"confidence"`
	AssignedAt      string                   `json:"assigned_at"`
	Recommendations []recommendationResponse `json:"recommendations"`
}

type recommendationResponse struct {
	ID       int    `json:"id"`
	Feature  string `json:"feature"`
	Reason   string `json:"reason"`
	Priority int    `json:"priority"`
}

func GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	var user userResponse

	error := utils.DB.QueryRow(`
		SELECT
			c.id,
			c.customer_id,
			c.email,
			COALESCE(cp.username, ''),
			COALESCE(cp.full_name, ''),
			c.created_at::text
		FROM customers c
		LEFT JOIN customer_profiles cp ON cp.customer_id = c.id
		WHERE c.id::text = $1 OR c.customer_id = $1
	`, id).Scan(
		&user.ID,
		&user.CustomerID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.CreatedAt,
	)

	if error != nil {
		if error == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "user found",
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

func GetAllUsers(ctx *gin.Context) {
	rows, error := utils.DB.Query(`
		SELECT
			c.id,
			c.customer_id,
			c.email,
			COALESCE(cp.username, ''),
			COALESCE(cp.full_name, ''),
			c.created_at::text
		FROM customers c
		LEFT JOIN customer_profiles cp ON cp.customer_id = c.id
		ORDER BY c.id ASC
	`)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}
	defer rows.Close()

	users := []userResponse{}

	for rows.Next() {
		var user userResponse
		if error := rows.Scan(
			&user.ID,
			&user.CustomerID,
			&user.Email,
			&user.Username,
			&user.FullName,
			&user.CreatedAt,
		); error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": error.Error(),
			})
			return
		}

		users = append(users, user)
	}

	if error := rows.Err(); error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "all users",
		"users":   users,
	})
}

func GetUserActivityById(ctx *gin.Context) {
	id := ctx.Param("id")

	var response struct {
		CustomerID     string  `json:"customer_id"`
		Transactions   []gin.H `json:"transactions"`
		FrequentlyUsed []gin.H `json:"frequently_used_features"`
	}

	err := utils.DB.QueryRow(`
		SELECT customer_id FROM customers WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&response.CustomerID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer transactionRows.Close()

	response.Transactions = []gin.H{}

	for transactionRows.Next() {
		var trxID, trxType, status, createdAt string
		var amount float64

		if err := transactionRows.Scan(&trxID, &trxType, &amount, &status, &createdAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response.Transactions = append(response.Transactions, gin.H{
			"trx_id":     trxID,
			"type":       trxType,
			"amount":     amount,
			"status":     status,
			"created_at": createdAt,
		})
	}

	if err := transactionRows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer featureRows.Close()

	response.FrequentlyUsed = []gin.H{}

	for featureRows.Next() {
		var feature, lastUsedAt string
		var usageCount int

		if err := featureRows.Scan(&feature, &usageCount, &lastUsedAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response.FrequentlyUsed = append(response.FrequentlyUsed, gin.H{
			"feature":      feature,
			"usage_count":  usageCount,
			"last_used_at": lastUsedAt,
		})
	}

	if err := featureRows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user activity found",
		"data":    response,
	})
}

func GetUserSegmentationById(ctx *gin.Context) {
	id := ctx.Param("id")

	var customerID string
	var customerInternalID int

	error := utils.DB.QueryRow(`
		SELECT id, customer_id
		FROM customers
		WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&customerInternalID, &customerID)

	if error != nil {
		if error == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	rows, error := utils.DB.Query(`
		SELECT
			s.id,
			s.name,
			COALESCE(s.description, ''),
			us.confidence,
			us.created_at::text
		FROM user_segments us
		JOIN segments s ON s.id = us.segment_id
		WHERE us.customer_id = $1
		ORDER BY us.confidence DESC, us.created_at DESC
	`, customerInternalID)
	if error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}
	defer rows.Close()

	segments := []segmentResponse{}

	for rows.Next() {
		var segment segmentResponse
		if error := rows.Scan(
			&segment.ID,
			&segment.Name,
			&segment.Description,
			&segment.Confidence,
			&segment.AssignedAt,
		); error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": error.Error(),
			})
			return
		}

		recommendations, error := getRecommendationsBySegmentID(segment.ID)
		if error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": error.Error(),
			})
			return
		}

		segment.Recommendations = recommendations
		segments = append(segments, segment)
	}

	if error := rows.Err(); error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(),
		})
		return
	}

	if len(segments) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "user segment not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "user segment found",
		"customer_id": customerID,
		"segments":    segments,
	})
}

func getRecommendationsBySegmentID(segmentID int) ([]recommendationResponse, error) {
	rows, error := utils.DB.Query(`
		SELECT
			id,
			COALESCE(feature, ''),
			COALESCE(reason, ''),
			priority
		FROM recommendations
		WHERE segment_id = $1
		ORDER BY priority ASC, id ASC
	`, segmentID)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	recommendations := []recommendationResponse{}

	for rows.Next() {
		var recommendation recommendationResponse
		if error := rows.Scan(
			&recommendation.ID,
			&recommendation.Feature,
			&recommendation.Reason,
			&recommendation.Priority,
		); error != nil {
			return nil, error
		}

		recommendations = append(recommendations, recommendation)
	}

	if error := rows.Err(); error != nil {
		return nil, error
	}

	return recommendations, nil
}
