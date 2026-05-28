package controllers

import (
	"B1K5-API/internal/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type personalizationResponse struct {
	CustomerID string `json:"customer_id"`
	Homepage   gin.H  `json:"homepage"`
}

func GetPersonalizationById(ctx *gin.Context) {
	id := ctx.Param("id")
	var response personalizationResponse
	var customerInternalID int

	// 1. cari user

	err := utils.DB.QueryRow(`
		SELECT id, customer_id
		FROM customers
		WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&customerInternalID, &response.CustomerID)

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

	// 2. ambil segment user

	var segmentID int
	var segmentName string
	var confidence float64

	err = utils.DB.QueryRow(`
		SELECT
			s.id,
			s.name,
			us.confidence
		FROM user_segments us
		JOIN segments s ON s.id = us.segment_id
		WHERE us.customer_id = $1
		ORDER BY us.confidence DESC, us.created_at DESC
		LIMIT 1
	`, customerInternalID).Scan(&segmentID, &segmentName, &confidence)

	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	segment := gin.H{}
	if err != sql.ErrNoRows {
		segment = gin.H{
			"id":         segmentID,
			"name":       segmentName,
			"confidence": confidence,
		}
	}

	// 3. ambil aktivitas/frequently used features

	rows, err := utils.DB.Query(`
		SELECT
			ua.feature,
			COUNT(*) AS usage_count,
			MAX(ua.created_at)::text AS last_used_at
		FROM user_activities ua
		WHERE ua.customer_id = $1
		GROUP BY ua.feature
		ORDER BY usage_count DESC, MAX(ua.created_at) DESC
		LIMIT 5
	`, customerInternalID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	quickActions := []gin.H{}
	primaryFeature := ""

	for rows.Next() {
		var feature, lastUsedAt string
		var usageCount int

		if err := rows.Scan(&feature, &usageCount, &lastUsedAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if primaryFeature == "" {
			primaryFeature = feature
		}

		quickActions = append(quickActions, gin.H{
			"feature":      feature,
			"usage_count":  usageCount,
			"last_used_at": lastUsedAt,
		})
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 4. susun konfigurasi homepage

	recommendations := []recommendationResponse{}
	if segmentID != 0 {
		recommendations, err = getRecommendationsBySegmentID(segmentID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if primaryFeature == "" && len(recommendations) > 0 {
			primaryFeature = recommendations[0].Feature
		}
	}

	response.Homepage = gin.H{
		"primary_feature":      primaryFeature,
		"segment":              segment,
		"quick_actions":        quickActions,
		"recommended_sections": recommendations,
	}

	// 5. return JSON

	ctx.JSON(http.StatusOK, gin.H{
		"message": "personalization config found",
		"data":    response,
	})
}
