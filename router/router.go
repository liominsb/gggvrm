package router // Package router 路由

import (
	"gggvrm/controllers"
	"gggvrm/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 依赖注入装配
	ctrls := Inject()

	r.Use(cors.New(cors.Config{
		// 允许哪些域来访问我？这里配置了前端的地址
		AllowOrigins: []string{"http://localhost:5173"}, // 在你的实际开发中，这里会改成 "http://localhost:5173"
		// 允许前端使用哪些危险方法？(解决预检请求问题)
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		// 允许前端携带哪些特殊的请求头？
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type"},
		// 允许前端读取哪些额外的响应头？
		ExposeHeaders: []string{"Content-Length"},
		// 是否允许携带 Cookie 等凭证？
		AllowCredentials: true,
		// 预检请求的结果缓存多久？(12小时内，同一个请求就不用再发 OPTIONS 探路了)
		MaxAge: 12 * time.Hour,
	}))

	r.Static("/uploads", "./uploads")

	auth := r.Group("/api/auth")
	{
		auth.POST("login", ctrls.AuthCtrl.Login)
		auth.POST("register", ctrls.AuthCtrl.Register)
		auth.POST("refreshTokens", ctrls.AuthCtrl.RefreshTokens)
	}

	api := r.Group("/api/v1")
	api.Use(middlewares.AuthMiddleware())
	{
		api.GET("/user", ctrls.AuthCtrl.Getmyuser)
		api.GET("/user/:id", ctrls.AuthCtrl.GetUserProfileById)
		api.PUT("/user", ctrls.AuthCtrl.Changepassword)

		api.POST("/article", ctrls.ArticleCtrl.CreateArticle)
		api.DELETE("/article/:id", ctrls.ArticleCtrl.DelArticle)
		api.GET("/articles", ctrls.ArticleCtrl.GetArticles)
		api.GET("/article/:id", ctrls.ArticleCtrl.GetArticlesByID)
		api.PUT("/article/:id", ctrls.ArticleCtrl.UpdateArticle)
		api.GET("/articles/cursor", ctrls.ArticleCtrl.GetArticlesByCursor)
		api.GET("/articles/feed", ctrls.FeedCtrl.GetUserFeed)

		api.POST("/article/:id/like", ctrls.LikeCtrl.LikeArticle)
		api.GET("/article/:id/like", ctrls.LikeCtrl.GetArticlelikes)

		api.POST("/article/:id/comment", ctrls.CommentCtrl.CreateComment)
		api.DELETE("/comment/:id", ctrls.CommentCtrl.DelComment)
		api.GET("/article/:id/comments", ctrls.CommentCtrl.GetComments)

		api.POST("/upload", controllers.UploadImage)

		api.GET("/tags", ctrls.TagsCtrl.GetTags)
		api.POST("/tag", ctrls.TagsCtrl.CreateTag)
		api.DELETE("/tag/:id", ctrls.TagsCtrl.DeleteTag)

		api.GET("/categories", ctrls.CateCtrl.GetCates)
		api.POST("/category", ctrls.CateCtrl.CreateCate)
		api.DELETE("/category/:id", ctrls.CateCtrl.DeleteCate)

		api.POST("/article/:id/favorite", ctrls.FavCtrl.ToggleFavorite)
		api.GET("/article/:id/favorite", ctrls.FavCtrl.GetFavoriteStatus)
		api.GET("/article/:id/favorites/count", ctrls.FavCtrl.GetFavoriteCount)
		api.GET("/user/favorites", ctrls.FavCtrl.GetUserFavorites)

		api.POST("/user/:id/follow", ctrls.FollowCtrl.ToggleFollow)
		api.GET("/user/:id/follow", ctrls.FollowCtrl.GetFollowStatus)
		api.GET("/user/:id/follow/counts", ctrls.FollowCtrl.GetFollowCounts)
		api.GET("/user/:id/following", ctrls.FollowCtrl.GetFollowing)
		api.GET("/user/:id/followers", ctrls.FollowCtrl.GetFollowers)

		api.GET("/ws", controllers.HandleConnections)
	}
	return r
}
