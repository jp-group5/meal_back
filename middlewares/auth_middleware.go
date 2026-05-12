package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"meal_back/models"
	"meal_back/pkg/auth"
	"meal_back/pkg/response"
	"meal_back/stores"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	codeUnauthorized = 10005
)

// AuthMiddleware JWT 鉴权中间件。
// 1. 从 Authorization 头中读取 Bearer Token。
// 2. 校验 token 合法性与 token 类型。
// 3. 校验 token 对应会话是否存在且未撤销（数据库持久化校验）。
// 4. 校验通过后把 userID、tokenJTI 等信息放入 Context。
func AuthMiddleware(db *gorm.DB, jwtSecret string, tokenBlacklist *stores.TokenBlacklistStore) gin.HandlerFunc {
	tokenService := auth.NewTokenService(jwtSecret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Missing Authorization header.")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Invalid Authorization format. Expected: Bearer <token>.")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Token cannot be empty.")
			c.Abort()
			return
		}

		claims, err := tokenService.ParseToken(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Invalid or expired token.")
			c.Abort()
			return
		}
		if claims.TokenType != auth.TokenTypeAccess {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Only access tokens are allowed.")
			c.Abort()
			return
		}
		if tokenBlacklist != nil && tokenBlacklist.IsBlacklisted(claims.ID) {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Token is revoked. Please log in again.")
			c.Abort()
			return
		}

		// 通过数据库校验会话状态，确保服务重启后依然能识别已撤销会话。
		var session models.UserSession
		if err := db.Where("access_token_jti = ? AND is_revoked = ?", claims.ID, false).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Session not found or revoked. Please log in again.")
				c.Abort()
				return
			}
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Failed to validate session.")
			c.Abort()
			return
		}

		if time.Now().After(session.AccessExpiresAt) {
			response.Error(c, http.StatusUnauthorized, codeUnauthorized, "Session expired. Refresh token or log in again.")
			c.Abort()
			return
		}

		// 更新最后活跃时间，失败不阻断业务请求。
		_ = db.Model(&models.UserSession{}).Where("id = ?", session.ID).Update("last_seen_at", time.Now()).Error

		c.Set("userID", claims.UserID)
		c.Set("tokenJTI", claims.ID)
		c.Set("tokenExpiresAt", claims.ExpiresAt.Unix())
		c.Set("sessionID", session.SessionID)
		c.Next()
	}
}
