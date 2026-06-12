package controllers

import (
	"database/sql"
	"net/http"

	"B1K5-API/internal/utils"

	"github.com/gin-gonic/gin"
)

type updateUserSegmentsInput struct {
	Segments []segmentUpdateInput `json:"segments"`
}

type segmentUpdateInput struct {
	CustomerID  string  `json:"customer_id"`
	SegmentName string  `json:"segment_name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type segmentUpdateResult struct {
	CustomerID  string  `json:"customer_id"`
	SegmentName string  `json:"segment_name"`
	Confidence  float64 `json:"confidence"`
	Action      string  `json:"action"`
}

func UpdateUserSegments(ctx *gin.Context) {
	var input updateUserSegmentsInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(input.Segments) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "segments is required",
		})
		return
	}

	tx, err := utils.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer tx.Rollback()

	results := []segmentUpdateResult{}

	for _, segment := range input.Segments {
		if segment.CustomerID == "" || segment.SegmentName == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "customer_id and segment_name are required",
			})
			return
		}

		if segment.Confidence < 0 || segment.Confidence > 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "confidence must be between 0 and 1",
			})
			return
		}

		customer, err := getCustomerByIDParamTx(tx, segment.CustomerID)
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

		segmentID, err := upsertSegment(tx, segment.SegmentName, segment.Description)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		action, err := upsertUserSegment(tx, customer.ID, segmentID, segment.Confidence)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		results = append(results, segmentUpdateResult{
			CustomerID:  customer.CustomerID,
			SegmentName: segment.SegmentName,
			Confidence:  segment.Confidence,
			Action:      action,
		})
	}

	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user segments updated",
		"updated": len(results),
		"results": results,
	})
}

func getCustomerByIDParamTx(tx *sql.Tx, id string) (customerIdentity, error) {
	var customer customerIdentity

	err := tx.QueryRow(`
		SELECT id, customer_id
		FROM customers
		WHERE id::text = $1 OR customer_id = $1
	`, id).Scan(&customer.ID, &customer.CustomerID)

	return customer, err
}

func upsertSegment(tx *sql.Tx, name string, description string) (int, error) {
	var segmentID int

	err := tx.QueryRow(`
		INSERT INTO segments (name, description)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET
			description = COALESCE(NULLIF(EXCLUDED.description, ''), segments.description)
		RETURNING id
	`, name, description).Scan(&segmentID)

	return segmentID, err
}

func upsertUserSegment(tx *sql.Tx, customerID int, segmentID int, confidence float64) (string, error) {
	result, err := tx.Exec(`
		UPDATE user_segments
		SET
			segment_id = $2,
			confidence = $3,
			created_at = CURRENT_TIMESTAMP
		WHERE customer_id = $1
	`, customerID, segmentID, confidence)
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected > 0 {
		return "updated", nil
	}

	_, err = tx.Exec(`
		INSERT INTO user_segments (customer_id, segment_id, confidence)
		VALUES ($1, $2, $3)
	`, customerID, segmentID, confidence)
	if err != nil {
		return "", err
	}

	return "inserted", nil
}
