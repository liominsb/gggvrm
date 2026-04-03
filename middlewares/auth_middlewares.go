package middlewares // Package middlewares 中间件

import (
	"gggvrm/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

//认证中间件

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			ctx.Abort()
			return
		}
		id, err := utils.ParseJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}
		ctx.Set("ID", id)
		ctx.Next()
	}
}
