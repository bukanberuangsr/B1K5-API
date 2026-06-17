package controllers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"

	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

type createAnalyticsEventInput struct {
	CustomerID string          `json:"customer_id"`
	EventType  string          `json:"event_type"`
	Feature    string          `json:"feature"`
	Metadata   json.RawMessage `json:"metadata"`
}

type analyticsMetricResponse struct {
	TotalEvents    int     `json:"total_events"`
	Impressions    int     `json:"impressions"`
	Clicks         int     `json:"clicks"`
	Engagements    int     `json:"engagements"`
	CTR            float64 `json:"ctr"`
	ConversionRate float64 `json:"conversion_rate"`
}

type featureMetricResponse struct {
	Feature    string `json:"feature"`
	EventCount int    `json:"event_count"`
	Clicks     int    `json:"clicks"`
}

func CreateAnalyticsEvent(ctx *gin.Context) {
	var input createAnalyticsEventInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if input.EventType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "event_type is required",
		})
		return
	}

	if input.Feature == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "feature is required",
		})
		return
	}

	tokenCustomerID, _ := ctx.Get("customer_id")
	role, _ := ctx.Get("role")

	if input.CustomerID == "" {
		input.CustomerID, _ = tokenCustomerID.(string)
	}

	if role != "admin" && input.CustomerID != tokenCustomerID {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
		return
	}

	customer, err := getCustomerByIDParam(input.CustomerID)
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

	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	if !json.Valid(metadata) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "metadata must be valid JSON",
		})
		return
	}

	var eventID int
	var createdAt string
	err = utils.DB.QueryRow(`
		INSERT INTO analytics_events (customer_id, event_type, feature, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at::text
	`, customer.ID, input.EventType, input.Feature, string(metadata)).Scan(&eventID, &createdAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "analytics event created",
		"event": gin.H{
			"id":          eventID,
			"customer_id": customer.CustomerID,
			"event_type":  input.EventType,
			"feature":     input.Feature,
			"metadata":    json.RawMessage(metadata),
			"created_at":  createdAt,
		},
	})
}

func GetAnalyticsMetrics(ctx *gin.Context) {
	metrics, err := getAnalyticsMetricSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	topFeatures, err := getTopFeatureMetrics()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "analytics metrics found",
		"metrics":      metrics,
		"top_features": topFeatures,
	})
}

func getAnalyticsMetricSummary() (analyticsMetricResponse, error) {
	var metrics analyticsMetricResponse

	err := utils.DB.QueryRow(`
		SELECT
			COUNT(*)::int AS total_events,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_impression', 'recommendation_viewed')
			)::int AS impressions,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
			)::int AS clicks,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_engaged', 'feature_used')
			)::int AS engagements
		FROM analytics_events
	`).Scan(
		&metrics.TotalEvents,
		&metrics.Impressions,
		&metrics.Clicks,
		&metrics.Engagements,
	)
	if err != nil {
		return metrics, err
	}

	if metrics.Impressions > 0 {
		metrics.CTR = math.Round((float64(metrics.Clicks)/float64(metrics.Impressions))*10000) / 100
	}

	if metrics.Clicks > 0 {
		metrics.ConversionRate = math.Round((float64(metrics.Engagements)/float64(metrics.Clicks))*10000) / 100
	}

	return metrics, nil
}

func getTopFeatureMetrics() ([]featureMetricResponse, error) {
	rows, err := utils.DB.Query(`
		SELECT
			COALESCE(feature, '') AS feature,
			COUNT(*)::int AS event_count,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
			)::int AS clicks
		FROM analytics_events
		GROUP BY feature
		ORDER BY event_count DESC, clicks DESC, feature ASC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	features := []featureMetricResponse{}

	for rows.Next() {
		var feature featureMetricResponse
		if err := rows.Scan(&feature.Feature, &feature.EventCount, &feature.Clicks); err != nil {
			return nil, err
		}

		features = append(features, feature)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return features, nil
}
