package utils

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func jwtSecretBytes() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return []byte("dev-insecure-secret-change-me")
	}
	return []byte(s)
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func GenerateJWT(userID uuid.UUID, role string) (string, error) {
	sec := jwtSecretBytes()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(sec)
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretBytes(), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		uid, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)
		if uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
			return
		}
		c.Set("userID", uid)
		c.Set("role", role)
		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not found in context"})
			return
		}
		role, _ := roleVal.(string)
		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}

func VerifiedOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("verified")
		if !ok || !v.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "email not verified"})
			return
		}
		c.Next()
	}
}

// OptionalJWTUserID returns (userID, true) if Authorization has valid Bearer JWT.
func OptionalJWTUserID(c *gin.Context) (string, bool) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		return "", false
	}
	tokenStr = strings.TrimPrefix(strings.TrimSpace(tokenStr), "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return "", false
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretBytes(), nil
	})
	if err != nil || !token.Valid {
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	uid, _ := claims["user_id"].(string)
	if uid == "" {
		return "", false
	}
	return uid, true
}
