package v1

import (
	"net/http"
	"os"

	"practice7/internal/entity"
	"practice7/internal/middleware"
	"practice7/internal/usecase"
	"practice7/pkg/logger"
	"practice7/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface) {
	r := &userRoutes{t: t, l: l}
	h := handler.Group("/users")
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)
		h.POST("/verify", r.VerifyEmail)

		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware())
		{
			protected.PATCH("/promote/:id", utils.RoleMiddleware("admin"), r.PromoteUser)

			verified := protected.Group("/")
			verified.Use(middleware.UserVerifiedContext(t), utils.VerifiedOnlyMiddleware())
			{
				verified.GET("/me", r.GetMe)
				verified.GET("/protected/hello", r.ProtectedHello)
			}
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var dto entity.CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := utils.HashPassword(dto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash password"})
		return
	}
	role := "user"
	if dto.Role == "admin" && os.Getenv("ALLOW_ADMIN_REGISTER") == "true" {
		role = "admin"
	}
	user := &entity.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: hashed,
		Role:     role,
	}
	created, sessionID, err := r.t.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered. Check email for the 4-digit verification code (or server logs if SMTP is off).",
		"session_id": sessionID,
		"user":       created,
	})
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (r *userRoutes) VerifyEmail(c *gin.Context) {
	var dto entity.VerifyEmailDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := r.t.VerifyEmail(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

func (r *userRoutes) GetMe(c *gin.Context) {
	uidStr := c.GetString("userID")
	id, err := uuid.Parse(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	u, err := r.t.GetMe(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"role":     u.Role,
		"verified": u.Verified,
	})
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	targetStr := c.Param("id")
	targetID, err := uuid.Parse(targetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	actorStr := c.GetString("userID")
	actorID, err := uuid.Parse(actorStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor"})
		return
	}
	if err := r.t.PromoteUser(actorID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user promoted to admin"})
}

func (r *userRoutes) ProtectedHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
