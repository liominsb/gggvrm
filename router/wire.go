package router

import (
	"gggvrm/controllers"
	"gggvrm/global"
	"gggvrm/repository"
	"gggvrm/service"
)

// Controllers 所有控制器的集合
type Controllers struct {
	LikeCtrl    *controllers.LikeController
	CommentCtrl *controllers.CommentController
	AuthCtrl    *controllers.AuthController
	ArticleCtrl *controllers.ArticleController
	TagsCtrl    *controllers.TagsController
	CateCtrl    *controllers.CateController
	FavCtrl     *controllers.FavoriteController
	FollowCtrl  *controllers.FollowController
	FeedCtrl    *controllers.FeedController
}

// Inject 依赖注入装配，像乐高积木一样一层层组装
func Inject() *Controllers {
	// Repository 层
	likeRepo := repository.NewLikeRepository(global.Db)
	commentRepo := repository.NewCommentRepository(global.Db)
	authRepo := repository.NewAuthRepository(global.Db)
	articleRepo := repository.NewArticleRepository(global.Db)
	tagsRepo := repository.NewTagsRepository(global.Db)
	cateRepo := repository.NewCateRepository(global.Db)
	favRepo := repository.NewFavoriteRepository(global.Db)
	followRepo := repository.NewFollowRepository(global.Db)

	// Service 层：拿到 Repo 和 Redis
	likeService := service.NewLikeService(likeRepo, global.RedisDB)
	commentService := service.NewCommentService(commentRepo, global.RedisDB)
	authService := service.NewAuthService(authRepo, global.RedisDB)
	articleService := service.NewArticleService(articleRepo, commentRepo, global.RedisDB)
	tagsService := service.NewTagsService(tagsRepo, global.RedisDB)
	cateService := service.NewCateService(cateRepo, global.RedisDB)
	favService := service.NewFavoriteService(favRepo, global.RedisDB)
	followService := service.NewFollowService(followRepo, global.RedisDB)
	feedService := service.NewFeedService(articleRepo, global.RedisDB)

	// Controller 层：拿到 Service
	return &Controllers{
		LikeCtrl:    controllers.NewLikeController(likeService),
		CommentCtrl: controllers.NewcCommentController(commentService),
		AuthCtrl:    controllers.NewAuthController(authService),
		ArticleCtrl: controllers.NewArticleController(articleService),
		TagsCtrl:    controllers.NewTagsController(tagsService),
		CateCtrl:    controllers.NewCateController(cateService),
		FavCtrl:     controllers.NewFavoriteController(favService),
		FollowCtrl:  controllers.NewFollowController(followService),
		FeedCtrl:    controllers.NewFeedController(feedService),
	}
}
