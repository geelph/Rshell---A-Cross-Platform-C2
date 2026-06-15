package middlewares

/*
修改说明：
1. BasicAuthMiddleware 添加用户存在判断。
2. Authorization2 添加合法性判断。
3. BasicAuthMiddleware 添加认证缓存，避免频繁查库和浏览器弹窗。
*/

import (
	"Rshell/pkg/common"
	"Rshell/pkg/database"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 认证缓存，避免每次请求都查数据库
var (
	authCache   = make(map[string]authCacheEntry)
	authCacheMu sync.RWMutex
)

const authCacheTTL = 10 * time.Minute

type authCacheEntry struct {
	user      string
	expiresAt time.Time
}

func getCachedAuth(authHeader string) (string, bool) {
	authCacheMu.RLock()
	defer authCacheMu.RUnlock()
	entry, ok := authCache[authHeader]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.user, true
}

func setCachedAuth(authHeader, user string) {
	authCacheMu.Lock()
	defer authCacheMu.Unlock()
	authCache[authHeader] = authCacheEntry{
		user:      user,
		expiresAt: time.Now().Add(authCacheTTL),
	}
}

// avoidPopup 对于静态资源不发送 WWW-Authenticate 头，避免浏览器弹出 Basic Auth 对话框
func avoidPopup(c *gin.Context) bool {
	path := c.Request.URL.Path
	return strings.HasPrefix(path, "/static/") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".svg") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".woff") ||
		strings.HasSuffix(path, ".woff2") ||
		strings.HasSuffix(path, ".ttf") ||
		strings.HasSuffix(path, ".eot")
}

func BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			// 静态资源和字体文件不发送 WWW-Authenticate，避免浏览器弹出登录框
			if !avoidPopup(c) {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 检查认证缓存
		if cachedUser, ok := getCachedAuth(authHeader); ok {
			c.Set("user", cachedUser)
			c.Next()
			return
		}

		encodedCreds := authHeader[len("Basic "):]
		creds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			if !avoidPopup(c) {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		credParts := strings.SplitN(string(creds), ":", 2)
		if len(credParts) != 2 {
			if !avoidPopup(c) {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, pass := credParts[0], credParts[1]

		var userPass database.Users
		has, err := database.Engine.Where("username = ?", user).Get(&userPass)
		if err != nil || !has {
			if !avoidPopup(c) {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if userPass.Password != pass || userPass.Password == "" {
			if !avoidPopup(c) {
				c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			}
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 缓存认证结果
		setCachedAuth(authHeader, user)

		c.Set("user", user)
		c.Next()
	}
}

// AuthMiddleware validates JWT from Authorization2 or query token.
func AuthMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                authHeader := strings.TrimSpace(c.GetHeader("Authorization2"))
                // 支持URL查询参数传递token，方便SSE之类的客户端
                if authHeader == "" {
                        queryToken := c.Query("token")
                        if queryToken != "" {
                                authHeader = "Bearer " + queryToken
                        }
                }

		if authHeader == "" {
			c.String(http.StatusUnauthorized, "Token required")
			c.Abort()
			return
		}

		if len(authHeader) < len("Bearer ") || !strings.EqualFold(authHeader[:len("Bearer ")], "Bearer ") {
			c.String(http.StatusUnauthorized, "Invalid token format")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
		if tokenString == "" {
			c.String(http.StatusUnauthorized, "Token required")
			c.Abort()
			return
		}

		claims, err := common.ValidateJWT(tokenString)
		if err != nil {
			// MCP SDK 期望标准的文本错误，避免抛出解析异常
			c.String(http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}
		c.Set("username", claims.Username)
		c.Next()
	}
}
