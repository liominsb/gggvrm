package router // Package router 路由

import (
	"gggvrm/controllers"
	"gggvrm/global"
	"gggvrm/middlewares"
	"gggvrm/repository"
	"gggvrm/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 2. 依赖注入装配 (像乐高积木一样一层层组装)
	likeRepo := repository.NewLikeRepository(global.Db)             // Repo 层只管自己
	likeService := service.NewLikeService(likeRepo, global.RedisDB) // Service 层拿到 Repo 和 Redis
	likeCtrl := controllers.NewLikeController(likeService)          // Controller 拿到 Service

	commentRepo := repository.NewCommentRepository(global.Db)
	commentService := service.NewCommentService(commentRepo, global.RedisDB)
	commentCtrl := controllers.NewcCommentController(commentService)

	authRepo := repository.NewAuthRepository(global.Db)
	authService := service.NewAuthService(authRepo, global.RedisDB)
	authCtrl := controllers.NewAuthController(authService)

	articleRepo := repository.NewArticleRepository(global.Db)
	articleService := service.NewArticleService(articleRepo, commentRepo, global.RedisDB)
	articleCtrl := controllers.NewArticleController(articleService)

	tagsRepo := repository.NewTagsRepository(global.Db)
	tagsService := service.NewTagsService(tagsRepo, global.RedisDB)
	tagsCtrl := controllers.NewTagsController(tagsService)

	cateRepo := repository.NewCateRepository(global.Db)
	cateService := service.NewCateService(cateRepo, global.RedisDB)
	cateCtrl := controllers.NewCateController(cateService)

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
		auth.POST("login", authCtrl.Login)
		auth.POST("register", authCtrl.Register)
		auth.POST("refreshTokens", authCtrl.RefreshTokens)
	}

	api := r.Group("/api/v1")
	api.Use(middlewares.AuthMiddleware())
	{
		api.GET("/user", authCtrl.Getmyuser)
		api.GET("/user/:id", authCtrl.GetUserProfileById)
		api.PUT("/user", authCtrl.Changepassword)

		api.POST("/article", articleCtrl.CreateArticle)
		api.DELETE("/article/:id", articleCtrl.DelArticle)
		api.GET("/articles", articleCtrl.GetArticles)
		api.GET("/article/:id", articleCtrl.GetArticlesByID)
		api.PUT("/article/:id", articleCtrl.UpdateArticle)
		api.GET("/articles/cursor", articleCtrl.GetArticlesByCursor)

		api.POST("/article/:id/like", likeCtrl.LikeArticle)
		api.GET("/article/:id/like", likeCtrl.GetArticlelikes)

		api.POST("/article/:id/comment", commentCtrl.CreateComment)
		api.DELETE("/comment/:id", commentCtrl.DelComment)
		api.GET("/article/:id/comments", commentCtrl.GetComments)

		api.POST("/upload", controllers.UploadImage)

		api.GET("/tags", tagsCtrl.GetTags)
		api.POST("/tag", tagsCtrl.CreateTag)
		api.DELETE("/tag/:id", tagsCtrl.DeleteTag)

		api.GET("/categories", cateCtrl.GetCates)
		api.POST("/category", cateCtrl.CreateCate)
		api.DELETE("/category/:id", cateCtrl.DeleteCate)

		api.GET("/ws", controllers.HandleConnections)
	}
	return r
}
