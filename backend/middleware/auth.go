package middleware

import (
	"net/http"
	"strings"
	"time"

	"lab-reservation-system/config"
	"lab-reservation-system/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint        `json:"user_id"`
	Username string      `json:"username"`
	RealName string      `json:"real_name"`
	Role     models.Role `json:"role"`
	GroupID  *uint       `json:"group_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RealName: user.RealName,
		Role:     user.Role,
		GroupID:  user.GroupID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "lab-reservation-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// AuthRequired JWT 鉴权中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请求未携带认证令牌Token"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token格式不正确，需为Bearer格式"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("real_name", claims.RealName)
		c.Set("role", claims.Role)
		c.Set("group_id", claims.GroupID)

		c.Next()
	}
}

// RequireRoles 角色权限检查中间件
func RequireRoles(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权限访问此接口"})
			c.Abort()
			return
		}

		userRole := roleVal.(models.Role)
		// 系统管理员拥有所有权限
		if userRole == models.RoleSystemAdmin {
			c.Next()
			return
		}

		hasRole := false
		for _, r := range roles {
			if userRole == r {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "当前角色权限不足，拒绝访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}
