package controllers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

// Dashboard Performance Response
type dashboardPerformanceResponse struct {
	OverallMetrics      metricsOverview        `json:"overall_metrics"`
	EngagementBySegment []engagementBySegment  `json:"engagement_by_segment"`
	ABTestResults       []abTestResult         `json:"ab_test_results"`
	RecommendationROI   recommendationROI      `json:"recommendation_roi"`
	TopPerformers       []topPerformingSegment `json:"top_performing_segments"`
}

type metricsOverview struct {
	TotalCustomers            int     `json:"total_customers"`
	ActiveCustomers           int     `json:"active_customers"`
	EngagementRate            float64 `json:"engagement_rate"`
	AvgRecommendationCTR      float64 `json:"avg_recommendation_ctr"`
	TotalRecommendations      int     `json:"total_recommendations"`
	SuccessfulRecommendations int     `json:"successful_recommendations"`
	DateRange                 string  `json:"date_range"`
}

type engagementBySegment struct {
	SegmentName               string  `json:"segment_name"`
	CustomerCount             int     `json:"customer_count"`
	ActiveCustomers           int     `json:"active_customers"`
	EngagementRate            float64 `json:"engagement_rate"`
	RecommendationImpressions int     `json:"recommendation_impressions"`
	RecommendationClicks      int     `json:"recommendation_clicks"`
	CTR                       float64 `json:"ctr"`
}

type abTestResult struct {
	TestID          int            `json:"test_id"`
	TestName        string         `json:"test_name"`
	Feature         string         `json:"feature"`
	VariantA        string         `json:"variant_a"`
	VariantB        string         `json:"variant_b"`
	VariantAMetrics variantMetrics `json:"variant_a_metrics"`
	VariantBMetrics variantMetrics `json:"variant_b_metrics"`
	WinningVariant  string         `json:"winning_variant"`
	StatisticalSig  float64        `json:"statistical_significance"`
	Status          string         `json:"status"`
}

type variantMetrics struct {
	CustomerCount  int     `json:"customer_count"`
	Engagements    int     `json:"engagements"`
	EngagementRate float64 `json:"engagement_rate"`
	Conversions    int     `json:"conversions"`
}

type recommendationROI struct {
	TotalRecommendations   int     `json:"total_recommendations"`
	ClickedRecommendations int     `json:"clicked_recommendations"`
	ROI                    float64 `json:"roi_percentage"`
	EstimatedValue         float64 `json:"estimated_value_idr"`
}

type topPerformingSegment struct {
	SegmentName    string  `json:"segment_name"`
	EngagementRate float64 `json:"engagement_rate"`
	CustomerCount  int     `json:"customer_count"`
	Rank           int     `json:"rank"`
}

// GET /api/dashboard/performance?days=30
func GetDashboardPerformance(ctx *gin.Context) {
	days := ctx.DefaultQuery("days", "30")

	overallMetrics, err := getMetricsOverview(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get overall metrics: %v", err),
		})
		return
	}

	engagementBySegment, err := getEngagementBySegment(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get engagement by segment: %v", err),
		})
		return
	}

	abTestResults, err := getActiveABTestResults(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get A/B test results: %v", err),
		})
		return
	}

	recommendationROI, err := getRecommendationROI(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get recommendation ROI: %v", err),
		})
		return
	}

	topPerformers, err := getTopPerformingSegments(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get top performers: %v", err),
		})
		return
	}

	response := dashboardPerformanceResponse{
		OverallMetrics:      overallMetrics,
		EngagementBySegment: engagementBySegment,
		ABTestResults:       abTestResults,
		RecommendationROI:   recommendationROI,
		TopPerformers:       topPerformers,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "dashboard performance data retrieved successfully",
		"data":    response,
	})
}

// Query Helper Functions

func getMetricsOverview(days string) (metricsOverview, error) {
	var metrics metricsOverview

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))
	dateRange := fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	err := utils.DB.QueryRow(`
		SELECT
			COUNT(DISTINCT c.id)::int AS total_customers,
			COUNT(DISTINCT CASE 
				WHEN ae.id IS NOT NULL THEN c.id 
			END)::int AS active_customers,
			COUNT(*) FILTER (
				WHERE ae.event_type IN ('recommendation_click', 'recommendation_clicked')
			)::int AS successful_recommendations,
			COUNT(*)::int AS total_recommendations
		FROM customers c
		LEFT JOIN analytics_events ae ON c.id = ae.customer_id 
			AND ae.created_at >= $1::timestamp
			AND ae.created_at < $2::timestamp
		WHERE c.role = 'customer'
	`, startDate, endDate).Scan(
		&metrics.TotalCustomers,
		&metrics.ActiveCustomers,
		&metrics.SuccessfulRecommendations,
		&metrics.TotalRecommendations,
	)

	if err != nil {
		return metrics, err
	}

	if metrics.TotalRecommendations > 0 {
		metrics.AvgRecommendationCTR = float64(metrics.SuccessfulRecommendations) / float64(metrics.TotalRecommendations)
	}

	if metrics.ActiveCustomers > 0 {
		metrics.EngagementRate = float64(metrics.ActiveCustomers) / float64(metrics.TotalCustomers)
	}

	metrics.DateRange = dateRange
	return metrics, nil
}

func getEngagementBySegment(days string) ([]engagementBySegment, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))

	rows, err := utils.DB.Query(`
		SELECT
			s.name,
			COUNT(DISTINCT us.customer_id)::int AS customer_count,
			COUNT(DISTINCT CASE 
				WHEN ae.id IS NOT NULL THEN us.customer_id 
			END)::int AS active_customers,
			COUNT(*) FILTER (
				WHERE ae.event_type IN ('recommendation_impression', 'recommendation_viewed')
			)::int AS impressions,
			COUNT(*) FILTER (
				WHERE ae.event_type IN ('recommendation_click', 'recommendation_clicked')
			)::int AS clicks
		FROM segments s
		LEFT JOIN user_segments us ON s.id = us.segment_id
		LEFT JOIN analytics_events ae ON us.customer_id = ae.customer_id
			AND ae.created_at >= $1::timestamp
			AND ae.created_at < $2::timestamp
		GROUP BY s.id, s.name
		ORDER BY customer_count DESC
	`, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []engagementBySegment
	for rows.Next() {
		var seg engagementBySegment
		if err := rows.Scan(
			&seg.SegmentName,
			&seg.CustomerCount,
			&seg.ActiveCustomers,
			&seg.RecommendationImpressions,
			&seg.RecommendationClicks,
		); err != nil {
			return nil, err
		}

		if seg.RecommendationImpressions > 0 {
			seg.CTR = float64(seg.RecommendationClicks) / float64(seg.RecommendationImpressions)
		}

		if seg.CustomerCount > 0 {
			seg.EngagementRate = float64(seg.ActiveCustomers) / float64(seg.CustomerCount)
		}

		results = append(results, seg)
	}

	return results, rows.Err()
}

func getActiveABTestResults(days string) ([]abTestResult, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))

	rows, err := utils.DB.Query(`
		SELECT
			t.id,
			t.test_name,
			t.feature,
			t.variant_a,
			t.variant_b,
			t.status,
			COUNT(CASE WHEN ata.variant = 'A' THEN 1 END)::int AS variant_a_customers,
			COUNT(CASE WHEN ata.variant = 'B' THEN 1 END)::int AS variant_b_customers,
			COUNT(DISTINCT CASE 
				WHEN ata.variant = 'A' AND ae.event_type IN ('recommendation_click', 'recommendation_clicked') 
				THEN ata.customer_id 
			END)::int AS variant_a_engagements,
			COUNT(DISTINCT CASE 
				WHEN ata.variant = 'B' AND ae.event_type IN ('recommendation_click', 'recommendation_clicked') 
				THEN ata.customer_id 
			END)::int AS variant_b_engagements
		FROM ab_tests t
		LEFT JOIN ab_test_assignments ata ON t.id = ata.ab_test_id
		LEFT JOIN analytics_events ae ON ata.customer_id = ae.customer_id 
			AND ae.created_at >= $1::timestamp
			AND ae.created_at < $2::timestamp
		WHERE t.status = 'active' OR (t.end_date >= $1::timestamp AND t.end_date < $2::timestamp)
		GROUP BY t.id, t.test_name, t.feature, t.variant_a, t.variant_b, t.status
	`, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []abTestResult
	for rows.Next() {
		var test abTestResult
		var variantACustomers, variantBCustomers int
		var variantAEngagements, variantBEngagements int

		if err := rows.Scan(
			&test.TestID,
			&test.TestName,
			&test.Feature,
			&test.VariantA,
			&test.VariantB,
			&test.Status,
			&variantACustomers,
			&variantBCustomers,
			&variantAEngagements,
			&variantBEngagements,
		); err != nil {
			return nil, err
		}

		test.VariantAMetrics.CustomerCount = variantACustomers
		test.VariantAMetrics.Engagements = variantAEngagements
		if variantACustomers > 0 {
			test.VariantAMetrics.EngagementRate = float64(variantAEngagements) / float64(variantACustomers)
		}

		test.VariantBMetrics.CustomerCount = variantBCustomers
		test.VariantBMetrics.Engagements = variantBEngagements
		if variantBCustomers > 0 {
			test.VariantBMetrics.EngagementRate = float64(variantBEngagements) / float64(variantBCustomers)
		}

		// Tentukan winning variant
		if test.VariantAMetrics.EngagementRate > test.VariantBMetrics.EngagementRate {
			test.WinningVariant = "A"
			test.StatisticalSig = (test.VariantAMetrics.EngagementRate - test.VariantBMetrics.EngagementRate) / test.VariantBMetrics.EngagementRate * 100
		} else {
			test.WinningVariant = "B"
			test.StatisticalSig = (test.VariantBMetrics.EngagementRate - test.VariantAMetrics.EngagementRate) / test.VariantAMetrics.EngagementRate * 100
		}

		results = append(results, test)
	}

	return results, rows.Err()
}

func getRecommendationROI(days string) (recommendationROI, error) {
	var roi recommendationROI
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))

	err := utils.DB.QueryRow(`
		SELECT
			COUNT(*)::int AS total_recommendations,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
			)::int AS clicked_recommendations
		FROM analytics_events
		WHERE created_at >= $1::timestamp
			AND created_at < $2::timestamp
			AND event_type LIKE 'recommendation%'
	`, startDate, endDate).Scan(
		&roi.TotalRecommendations,
		&roi.ClickedRecommendations,
	)

	if err != nil {
		return roi, err
	}

	if roi.TotalRecommendations > 0 {
		roi.ROI = (float64(roi.ClickedRecommendations) / float64(roi.TotalRecommendations)) * 100
		// Asumsi: setiap click recommendation = Rp 10,000 value
		roi.EstimatedValue = float64(roi.ClickedRecommendations) * 10000
	}

	return roi, nil
}

func getTopPerformingSegments(days string) ([]topPerformingSegment, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))

	rows, err := utils.DB.Query(`
		SELECT
			s.name,
			(COUNT(DISTINCT CASE 
				WHEN ae.id IS NOT NULL THEN us.customer_id 
			END)::float / NULLIF(COUNT(DISTINCT us.customer_id), 0)) AS engagement_rate,
			COUNT(DISTINCT us.customer_id)::int AS customer_count,
			ROW_NUMBER() OVER (ORDER BY COUNT(DISTINCT CASE 
				WHEN ae.id IS NOT NULL THEN us.customer_id 
			END)::float / NULLIF(COUNT(DISTINCT us.customer_id), 0) DESC) AS rank
		FROM segments s
		LEFT JOIN user_segments us ON s.id = us.segment_id
		LEFT JOIN analytics_events ae ON us.customer_id = ae.customer_id
			AND ae.created_at >= $1::timestamp
			AND ae.created_at < $2::timestamp
		GROUP BY s.id, s.name
		ORDER BY rank
		LIMIT 5
	`, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []topPerformingSegment
	for rows.Next() {
		var seg topPerformingSegment
		if err := rows.Scan(
			&seg.SegmentName,
			&seg.EngagementRate,
			&seg.CustomerCount,
			&seg.Rank,
		); err != nil {
			return nil, err
		}
		results = append(results, seg)
	}

	return results, rows.Err()
}

func parseDays(daysStr string) int {
	days := 30
	if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
		days = 30
	}
	return days
}

// Engagement Dashboard

type engagementDashboardResponse struct {
	CTR                    float64                `json:"ctr"`
	ConversionRate         float64                `json:"conversion_rate"`
	TotalClicksThisWeek    int                    `json:"total_clicks_this_week"`
	TopRecommendedFeatures []topRecommendedFeature `json:"top_recommended_features"`
	Period                 string                 `json:"period"`
}

type topRecommendedFeature struct {
	Rank        int    `json:"rank"`
	Feature     string `json:"feature"`
	TotalClicks int    `json:"total_clicks"`
}

// GET /api/dashboard/engagement
func GetEngagementDashboard(ctx *gin.Context) {
	var impressions, totalClicks, engagements, weeklyClicks int

	err := utils.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_impression', 'recommendation_viewed'))::int,
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_click', 'recommendation_clicked'))::int,
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_engaged', 'feature_used'))::int,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
				  AND created_at >= NOW() - INTERVAL '7 days'
			)::int
		FROM analytics_events
	`).Scan(&impressions, &totalClicks, &engagements, &weeklyClicks)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var ctr float64
	if impressions > 0 {
		ctr = math.Round((float64(totalClicks)/float64(impressions))*10000) / 100
	}

	var conversionRate float64
	if totalClicks > 0 {
		conversionRate = math.Round((float64(engagements)/float64(totalClicks))*10000) / 100
	}

	rows, err := utils.DB.Query(`
		SELECT
			feature,
			COUNT(*)::int AS total_clicks,
			ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC)::int AS rank
		FROM analytics_events
		WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
		  AND feature IS NOT NULL AND feature != ''
		GROUP BY feature
		ORDER BY total_clicks DESC
		LIMIT 3
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	topFeatures := []topRecommendedFeature{}
	for rows.Next() {
		var f topRecommendedFeature
		if err := rows.Scan(&f.Feature, &f.TotalClicks, &f.Rank); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		topFeatures = append(topFeatures, f)
	}
	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "engagement dashboard data retrieved successfully",
		"data": engagementDashboardResponse{
			CTR:                    ctr,
			ConversionRate:         conversionRate,
			TotalClicksThisWeek:    weeklyClicks,
			TopRecommendedFeatures: topFeatures,
			Period:                 fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		},
	})
}
