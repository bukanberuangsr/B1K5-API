package controllers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

// ─── Segment 1 & 4: Metrics ───────────────────────────────────────────────────

type metricWithChange struct {
	Value      float64 `json:"value"`
	LastWeek   float64 `json:"last_week"`
	ChangeRate float64 `json:"change_rate"` // persen perubahan dari minggu lalu
}

type dashboardMetricsResponse struct {
	CTR            metricWithChange `json:"ctr"`
	ConversionRate metricWithChange `json:"conversion_rate"`
	TotalClicks    metricWithChange `json:"total_clicks"`
	Segment        string           `json:"segment,omitempty"` // kosong = semua segment
}

// GET /api/dashboard/metrics?segment=Digital+Spender
func GetDashboardMetrics(ctx *gin.Context) {
	segment := ctx.Query("segment") // opsional

	now := time.Now()
	thisWeekStart := now.AddDate(0, 0, -7)
	lastWeekStart := now.AddDate(0, 0, -14)

	// Bangun kondisi JOIN segment jika filter aktif
	segmentJoin := ""
	segmentWhere := ""
	args := []interface{}{thisWeekStart, now, lastWeekStart, thisWeekStart}
	if segment != "" {
		segmentJoin = `
			JOIN LATERAL (
				SELECT segment_id
				FROM user_segments
				WHERE customer_id = ae.customer_id
				ORDER BY confidence DESC, created_at DESC
				LIMIT 1
			) us ON true
			JOIN segments s ON us.segment_id = s.id`
		segmentWhere = fmt.Sprintf(" AND LOWER(REPLACE(s.name, '_', ' ')) = LOWER(REPLACE($%d, '_', ' '))", len(args)+1)
		args = append(args, segment)
	}

	query := fmt.Sprintf(`
		SELECT
			-- minggu ini
			COUNT(*) FILTER (WHERE ae.created_at >= $1 AND ae.created_at < $2
				AND ae.event_type IN ('recommendation_impression','recommendation_viewed'))::int AS imp_now,
			COUNT(*) FILTER (WHERE ae.created_at >= $1 AND ae.created_at < $2
				AND ae.event_type IN ('recommendation_click','recommendation_clicked'))::int  AS click_now,
			COUNT(*) FILTER (WHERE ae.created_at >= $1 AND ae.created_at < $2
				AND ae.event_type IN ('recommendation_engaged','feature_used'))::int           AS eng_now,
			-- minggu lalu
			COUNT(*) FILTER (WHERE ae.created_at >= $3 AND ae.created_at < $4
				AND ae.event_type IN ('recommendation_impression','recommendation_viewed'))::int AS imp_last,
			COUNT(*) FILTER (WHERE ae.created_at >= $3 AND ae.created_at < $4
				AND ae.event_type IN ('recommendation_click','recommendation_clicked'))::int  AS click_last,
			COUNT(*) FILTER (WHERE ae.created_at >= $3 AND ae.created_at < $4
				AND ae.event_type IN ('recommendation_engaged','feature_used'))::int           AS eng_last
		FROM analytics_events ae
		%s
		WHERE 1=1%s
	`, segmentJoin, segmentWhere)

	var impNow, clickNow, engNow, impLast, clickLast, engLast int
	err := utils.DB.QueryRow(query, args...).Scan(
		&impNow, &clickNow, &engNow,
		&impLast, &clickLast, &engLast,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctrNow := calcCTR(clickNow, impNow)
	ctrLast := calcCTR(clickLast, impLast)

	convNow := calcCTR(engNow, clickNow)
	convLast := calcCTR(engLast, clickLast)

	resp := dashboardMetricsResponse{
		CTR: metricWithChange{
			Value:      ctrNow,
			LastWeek:   ctrLast,
			ChangeRate: changeRate(ctrNow, ctrLast),
		},
		ConversionRate: metricWithChange{
			Value:      convNow,
			LastWeek:   convLast,
			ChangeRate: changeRate(convNow, convLast),
		},
		TotalClicks: metricWithChange{
			Value:      float64(clickNow),
			LastWeek:   float64(clickLast),
			ChangeRate: changeRate(float64(clickNow), float64(clickLast)),
		},
		Segment: segment,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "dashboard metrics retrieved successfully",
		"data":    resp,
	})
}

// ─── Segment 2: Daily Clicks Chart ───────────────────────────────────────────

type dailyClicksResponse struct {
	Labels []string  `json:"labels"` // ["Senin","Selasa",...]
	Values []int     `json:"values"` // clicks per hari
	Period string    `json:"period"`
}

// GET /api/dashboard/clicks-chart
func GetWeeklyClicksChart(ctx *gin.Context) {
	now := time.Now()
	// Mulai dari Senin minggu ini
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	rows, err := utils.DB.Query(`
		SELECT
			DATE(created_at) AS day,
			COUNT(*)::int    AS clicks
		FROM analytics_events
		WHERE event_type IN ('recommendation_click', 'recommendation_clicked')
		  AND created_at >= $1
		  AND created_at <  $2
		GROUP BY day
		ORDER BY day
	`, monday, monday.AddDate(0, 0, 7))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Buat map hari → clicks
	dayNames := []string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}
	clicksByDate := map[string]int{}
	for rows.Next() {
		var day time.Time
		var clicks int
		if err := rows.Scan(&day, &clicks); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		clicksByDate[day.Format("2006-01-02")] = clicks
	}
	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	labels := make([]string, 7)
	values := make([]int, 7)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		labels[i] = dayNames[i]
		values[i] = clicksByDate[d.Format("2006-01-02")]
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "weekly clicks chart retrieved successfully",
		"data": dailyClicksResponse{
			Labels: labels,
			Values: values,
			Period: fmt.Sprintf("%s s/d %s", monday.Format("2006-01-02"), monday.AddDate(0, 0, 6).Format("2006-01-02")),
		},
	})
}

// ─── Segment 3: Top 3 Fitur Populer ──────────────────────────────────────────

type topFeatureItem struct {
	Rank        int    `json:"rank"`
	Feature     string `json:"feature"`
	TotalClicks int    `json:"total_clicks"`
}

// GET /api/dashboard/top-features
func GetTopFeatures(ctx *gin.Context) {
	rows, err := utils.DB.Query(`
		SELECT
			feature,
			COUNT(*)::int                              AS total_clicks,
			ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) AS rank
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

	features := []topFeatureItem{}
	for rows.Next() {
		var f topFeatureItem
		if err := rows.Scan(&f.Feature, &f.TotalClicks, &f.Rank); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		features = append(features, f)
	}
	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "top features retrieved successfully",
		"data":    features,
	})
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func calcCTR(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return math.Round(float64(num)/float64(den)*10000) / 100 // persen, 2 desimal
}

func changeRate(now, last float64) float64 {
	if last == 0 {
		if now == 0 {
			return 0
		}
		return 100 // dari 0 ke ada = +100%
	}
	return math.Round((now-last)/last*10000) / 100 // persen, 2 desimal
}

// ─── Legacy endpoints (tetap ada untuk kompatibilitas) ────────────────────────

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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get overall metrics: %v", err)})
		return
	}

	engagementBySegment, err := getEngagementBySegment(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get engagement by segment: %v", err)})
		return
	}

	abTestResults, err := getActiveABTestResults(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get A/B test results: %v", err)})
		return
	}

	recommendationROI, err := getRecommendationROI(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get recommendation ROI: %v", err)})
		return
	}

	topPerformers, err := getTopPerformingSegments(days)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get top performers: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "dashboard performance data retrieved successfully",
		"data": dashboardPerformanceResponse{
			OverallMetrics:      overallMetrics,
			EngagementBySegment: engagementBySegment,
			ABTestResults:       abTestResults,
			RecommendationROI:   recommendationROI,
			TopPerformers:       topPerformers,
		},
	})
}

// GET /api/dashboard/engagement (legacy)
type engagementDashboardResponse struct {
	CTR                    float64          `json:"ctr"`
	ConversionRate         float64          `json:"conversion_rate"`
	TotalClicksThisWeek    int              `json:"total_clicks_this_week"`
	TopRecommendedFeatures []topFeatureItem `json:"top_recommended_features"`
	Period                 string           `json:"period"`
}

func GetEngagementDashboard(ctx *gin.Context) {
	var impressions, totalClicks, engagements, weeklyClicks int
	err := utils.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_impression','recommendation_viewed'))::int,
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_click','recommendation_clicked'))::int,
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_engaged','feature_used'))::int,
			COUNT(*) FILTER (
				WHERE event_type IN ('recommendation_click','recommendation_clicked')
				  AND created_at >= NOW() - INTERVAL '7 days'
			)::int
		FROM analytics_events
	`).Scan(&impressions, &totalClicks, &engagements, &weeklyClicks)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctr := calcCTR(totalClicks, impressions)
	conversionRate := calcCTR(engagements, totalClicks)

	rows, err := utils.DB.Query(`
		SELECT feature, COUNT(*)::int, ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC)::int
		FROM analytics_events
		WHERE event_type IN ('recommendation_click','recommendation_clicked')
		  AND feature IS NOT NULL AND feature != ''
		GROUP BY feature
		ORDER BY 2 DESC
		LIMIT 3
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	topFeatures := []topFeatureItem{}
	for rows.Next() {
		var f topFeatureItem
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

// ─── Internal query helpers ───────────────────────────────────────────────────

func parseDays(daysStr string) int {
	days := 30
	fmt.Sscanf(daysStr, "%d", &days)
	return days
}

func getMetricsOverview(days string) (metricsOverview, error) {
	var metrics metricsOverview
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -parseDays(days))
	dateRange := fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	err := utils.DB.QueryRow(`
		SELECT
			COUNT(DISTINCT c.id)::int,
			COUNT(DISTINCT CASE WHEN ae.id IS NOT NULL THEN c.id END)::int,
			COUNT(*) FILTER (WHERE ae.event_type IN ('recommendation_click','recommendation_clicked'))::int,
			COUNT(*)::int
		FROM customers c
		LEFT JOIN analytics_events ae ON c.id = ae.customer_id
			AND ae.created_at >= $1::timestamp
			AND ae.created_at <  $2::timestamp
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
	if metrics.TotalCustomers > 0 {
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
			COUNT(DISTINCT us.customer_id)::int,
			COUNT(DISTINCT CASE WHEN ae.id IS NOT NULL THEN us.customer_id END)::int,
			COUNT(*) FILTER (WHERE ae.event_type IN ('recommendation_impression','recommendation_viewed'))::int,
			COUNT(*) FILTER (WHERE ae.event_type IN ('recommendation_click','recommendation_clicked'))::int
		FROM segments s
		LEFT JOIN user_segments us ON s.id = us.segment_id
		LEFT JOIN analytics_events ae ON us.customer_id = ae.customer_id
			AND ae.created_at >= $1::timestamp
			AND ae.created_at <  $2::timestamp
		GROUP BY s.id, s.name
		ORDER BY 2 DESC
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
			t.id, t.test_name, t.feature, t.variant_a, t.variant_b, t.status,
			COUNT(CASE WHEN ata.variant = 'A' THEN 1 END)::int,
			COUNT(CASE WHEN ata.variant = 'B' THEN 1 END)::int,
			COUNT(DISTINCT CASE WHEN ata.variant = 'A' AND ae.event_type IN ('recommendation_click','recommendation_clicked') THEN ata.customer_id END)::int,
			COUNT(DISTINCT CASE WHEN ata.variant = 'B' AND ae.event_type IN ('recommendation_click','recommendation_clicked') THEN ata.customer_id END)::int
		FROM ab_tests t
		LEFT JOIN ab_test_assignments ata ON t.id = ata.ab_test_id
		LEFT JOIN analytics_events ae ON ata.customer_id = ae.customer_id
			AND ae.created_at >= $1::timestamp AND ae.created_at < $2::timestamp
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
		var aCount, bCount, aEng, bEng int
		if err := rows.Scan(&test.TestID, &test.TestName, &test.Feature, &test.VariantA, &test.VariantB, &test.Status, &aCount, &bCount, &aEng, &bEng); err != nil {
			return nil, err
		}
		test.VariantAMetrics = variantMetrics{CustomerCount: aCount, Engagements: aEng}
		test.VariantBMetrics = variantMetrics{CustomerCount: bCount, Engagements: bEng}
		if aCount > 0 {
			test.VariantAMetrics.EngagementRate = float64(aEng) / float64(aCount)
		}
		if bCount > 0 {
			test.VariantBMetrics.EngagementRate = float64(bEng) / float64(bCount)
		}
		if test.VariantAMetrics.EngagementRate >= test.VariantBMetrics.EngagementRate {
			test.WinningVariant = "A"
		} else {
			test.WinningVariant = "B"
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
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE event_type IN ('recommendation_click','recommendation_clicked'))::int
		FROM analytics_events
		WHERE created_at >= $1::timestamp AND created_at < $2::timestamp
		  AND event_type LIKE 'recommendation%'
	`, startDate, endDate).Scan(&roi.TotalRecommendations, &roi.ClickedRecommendations)
	if err != nil {
		return roi, err
	}
	if roi.TotalRecommendations > 0 {
		roi.ROI = float64(roi.ClickedRecommendations) / float64(roi.TotalRecommendations) * 100
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
			(COUNT(DISTINCT CASE WHEN ae.id IS NOT NULL THEN us.customer_id END)::float
				/ NULLIF(COUNT(DISTINCT us.customer_id), 0)) AS engagement_rate,
			COUNT(DISTINCT us.customer_id)::int,
			ROW_NUMBER() OVER (ORDER BY
				COUNT(DISTINCT CASE WHEN ae.id IS NOT NULL THEN us.customer_id END)::float
				/ NULLIF(COUNT(DISTINCT us.customer_id), 0) DESC
			) AS rank
		FROM segments s
		LEFT JOIN user_segments us ON s.id = us.segment_id
		LEFT JOIN analytics_events ae ON us.customer_id = ae.customer_id
			AND ae.created_at >= $1::timestamp AND ae.created_at < $2::timestamp
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
		if err := rows.Scan(&seg.SegmentName, &seg.EngagementRate, &seg.CustomerCount, &seg.Rank); err != nil {
			return nil, err
		}
		results = append(results, seg)
	}
	return results, rows.Err()
}
