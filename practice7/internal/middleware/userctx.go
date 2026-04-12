package middleware

import (
	"net/http"

	"practice7/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UserVerifiedContext(uc usecase.UserInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidStr, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		s, _ := uidStr.(string)
		id, err := uuid.Parse(s)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}
		verified, err := uc.UserVerified(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.Set("verified", verified)
		c.Next()
	}
}
