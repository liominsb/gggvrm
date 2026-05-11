package middlewares // Package middlewares 中间件

import (
	"fmt"
	"gggvrm/global"
	"gggvrm/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

//认证中间件

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"错误": "未授权"})
			ctx.Abort()
			return
		}
		token = token[7:]

		claims, err := utils.ParseToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"错误": "无效令牌"})
			ctx.Abort()
			return
		}

		// 防踢人下线校验：查询 Redis 中的合法 SessionID
		redisKey := fmt.Sprintf("auth:account:%d", claims.AccountID)
		activeSessionID, err := global.RedisDB.HGet(ctx.Request.Context(), redisKey, "session_id").Result()

		if err == redis.Nil {
			// Redis 中找不到记录，说明登录已过期或被后台强制清除
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"错误": "会话已结束或终止"})
			return
		} else if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"错误": "内部服务器错误"})
			return
		}

		if claims.SessionID != activeSessionID {
			// Token 里的 SessionID 与服务器当前的 SessionID 不一致
			// 说明该账号已经在其他设备登录，当前设备被踢下线！
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"error": "kicked_out",
				"msg":   "您已在其他设备登录，当前设备被迫下线",
			})
			return
		}

		ctx.Set("ID", claims.AccountID)
		ctx.Next()
	}
}
