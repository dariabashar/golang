package v1

import (
	"practice7/internal/usecase"
	"practice7/pkg/logger"
	"practice7/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, uc usecase.UserInterface, l logger.Interface) {
	v1 := r.Group("/api/v1")
	v1.Use(utils.RateLimitMiddleware())
	{
		newUserRoutes(v1, uc, l)
	}
}
