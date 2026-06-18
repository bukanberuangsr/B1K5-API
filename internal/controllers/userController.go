package controllers

import (
	"B1K5-API/internal/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userResponse struct {
	ID                        int    `json:"id"`
	CustomerID                string `json:"customer_id"`
	Email                     string `json:"email"`
	Username                  string `json:"username"`
	FullName                  string `json:"full_name"`
	Role                      string `json:"role"`
	IsActivityTrackingEnabled bool   `json:"is_activity_tracking_enabled"`
	CreatedAt                 string `json:"created_at"`
}

type userWithSegmentResponse struct {
	ID                        int     `json:"id"`
	CustomerID                string  `json:"customer_id"`
	Email                     string  `json:"email"`
	Username                  string  `json:"username"`
	FullName                  string  `json:"full_name"`
	Role                      string  `json:"role"`
	IsActivityTrackingEnabled bool    `json:"is_activity_tracking_enabled"`
	CreatedAt                 string  `json:"created_at"`
	Segment                   string  `json:"segment"`       // nama segment, kosong jika belum ada
	Confidence                float64 `json:"confidence"`    // 0 jika belum ada segment
	AssignedAt                string  `json:"assigned_at"`   // kosong jika belum ada
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
			COALESCE(c.username, ''),
			COALESCE(cp.full_name, ''),
			c.role,
			c.is_activity_tracking_enabled,
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
		&user.Role,
		&user.IsActivityTrackingEnabled,
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
		"message":                      "user found",
		"email":                        user.Email,
		"username":                     user.Username,
		"role":                         user.Role,
		"is_activity_tracking_enabled": user.IsActivityTrackingEnabled,
		"created_at":                   user.CreatedAt,
	})
}

func GetAllUsers(ctx *gin.Context) {
	rows, error := utils.DB.Query(`
		SELECT
			c.id,
			c.customer_id,
			c.email,
			COALESCE(c.username, ''),
			COALESCE(cp.full_name, ''),
			c.role,
			c.is_activity_tracking_enabled,
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
			&user.Role,
			&user.IsActivityTrackingEnabled,
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

// GetAllUsersWithSegments mengembalikan semua user beserta segment aktif mereka.
// Segment diambil berdasarkan confidence tertinggi.
// GET /api/users/with-segments
func GetAllUsersWithSegments(ctx *gin.Context) {
	rows, err := utils.DB.Query(`
		SELECT
			c.id,
			c.customer_id,
			c.email,
			COALESCE(c.username, ''),
			COALESCE(cp.full_name, ''),
			c.role,
			c.is_activity_tracking_enabled,
			c.created_at::text,
			COALESCE(s.name, '')          AS segment,
			COALESCE(us.confidence, 0)    AS confidence,
			COALESCE(us.created_at::text, '') AS assigned_at
		FROM customers c
		LEFT JOIN customer_profiles cp ON cp.customer_id = c.id
		LEFT JOIN LATERAL (
			SELECT segment_id, confidence, created_at
			FROM user_segments
			WHERE customer_id = c.id
			ORDER BY confidence DESC, created_at DESC
			LIMIT 1
		) us ON true
		LEFT JOIN segments s ON s.id = us.segment_id
		ORDER BY c.id ASC
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	users := []userWithSegmentResponse{}
	for rows.Next() {
		var u userWithSegmentResponse
		if err := rows.Scan(
			&u.ID,
			&u.CustomerID,
			&u.Email,
			&u.Username,
			&u.FullName,
			&u.Role,
			&u.IsActivityTrackingEnabled,
			&u.CreatedAt,
			&u.Segment,
			&u.Confidence,
			&u.AssignedAt,
		); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "all users with segments",
		"total":   len(users),
		"users":   users,
	})
}

func GetUserActivityById(ctx *gin.Context) {
	id := ctx.Param("id")

	data, err := getUserActivityData(id)
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

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user activity found",
		"data":    data,
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

type updateTrackingInput struct {
	IsActivityTrackingEnabled *bool `json:"is_activity_tracking_enabled"`
}

// PUT /api/users/:id/tracking
func UpdateActivityTracking(ctx *gin.Context) {
	id := ctx.Param("id")

	var input updateTrackingInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if input.IsActivityTrackingEnabled == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "is_activity_tracking_enabled is required",
		})
		return
	}

	result, err := utils.DB.Exec(`
		UPDATE customers
		SET is_activity_tracking_enabled = $1
		WHERE id::text = $2 OR customer_id = $2
	`, *input.IsActivityTrackingEnabled, id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":                      "activity tracking updated successfully",
		"is_activity_tracking_enabled": *input.IsActivityTrackingEnabled,
	})
}
