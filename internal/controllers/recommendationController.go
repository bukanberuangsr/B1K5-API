package controllers

import (
	"database/sql"
	"net/http"

	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

type userRecommendationResponse struct {
	ID         int     `json:"id,omitempty"`
	Feature    string  `json:"feature"`
	Reason     string  `json:"reason"`
	Priority   int     `json:"priority"`
	Source     string  `json:"source"`
	Segment    string  `json:"segment,omitempty"`
	UsageCount int     `json:"usage_count,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

func GetRecommendationByUserID(ctx *gin.Context) {
	customer, err := getCustomerByIDParam(ctx.Param("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	recommendations, err := getRecommendationsFromSegments(customer.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(recommendations) == 0 {
		recommendations, err = getRecommendationsFromActivities(customer.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "recommendations found",
		"customer_id":     customer.CustomerID,
		"recommendations": recommendations,
	})
}

func getRecommendationsFromSegments(customerID int) ([]userRecommendationResponse, error) {
	rows, err := utils.DB.Query(`
		SELECT
			r.id,
			COALESCE(r.feature, ''),
			COALESCE(r.reason, ''),
			r.priority,
			s.name,
			us.confidence
		FROM user_segments us
		JOIN segments s ON s.id = us.segment_id
		JOIN recommendations r ON r.segment_id = s.id
		WHERE us.customer_id = $1
		ORDER BY us.confidence DESC, r.priority ASC, r.id ASC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recommendations := []userRecommendationResponse{}

	for rows.Next() {
		var recommendation userRecommendationResponse
		if err := rows.Scan(
			&recommendation.ID,
			&recommendation.Feature,
			&recommendation.Reason,
			&recommendation.Priority,
			&recommendation.Segment,
			&recommendation.Confidence,
		); err != nil {
			return nil, err
		}

		recommendation.Source = "segment"
		recommendations = append(recommendations, recommendation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recommendations, nil
}

func getRecommendationsFromActivities(customerID int) ([]userRecommendationResponse, error) {
	rows, err := utils.DB.Query(`
		SELECT
			ua.feature,
			COUNT(*) AS usage_count
		FROM user_activities ua
		WHERE ua.customer_id = $1
		GROUP BY ua.feature
		ORDER BY usage_count DESC, MAX(ua.created_at) DESC
		LIMIT 5
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recommendations := []userRecommendationResponse{}
	priority := 1

	for rows.Next() {
		var feature string
		var usageCount int

		if err := rows.Scan(&feature, &usageCount); err != nil {
			return nil, err
		}

		recommendations = append(recommendations, userRecommendationResponse{
			Feature:    feature,
			Reason:     "Fitur ini relevan karena sering digunakan oleh user.",
			Priority:   priority,
			Source:     "activity",
			UsageCount: usageCount,
		})
		priority++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recommendations, nil
}
