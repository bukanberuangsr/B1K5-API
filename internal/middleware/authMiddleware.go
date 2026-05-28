package middleware

import (
	"B1K5-API/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must use Bearer token",
			})
			return
		}

		claims, err := utils.ParseJWT(parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("customer_id", claims.CustomerID)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, ok := ctx.Get("role")
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authenticated role is required",
			})
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
	}
}

func RequireSelfOrRole(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, _ := ctx.Get("role")
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				ctx.Next()
				return
			}
		}

		paramID := ctx.Param("id")
		userID, _ := ctx.Get("user_id")
		customerID, _ := ctx.Get("customer_id")

		if paramID == customerID {
			ctx.Next()
			return
		}

		if userIDInt, ok := userID.(int); ok && paramID == strconv.Itoa(userIDInt) {
			ctx.Next()
			return
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "forbidden",
		})
	}
}
