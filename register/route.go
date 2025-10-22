package register

import (
	"app/api/block"
	"app/api/confession"
	"app/api/user"
	"github.com/zjutjh/mygo/jwt/middleware"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		userGroup := r.Group("/user")
		{
			userGroup.POST("/upload_image", middleware.Auth(true), api.UploadImageHandler())
			userGroup.POST("/register", user.RegisterHandler())
			userGroup.POST("/login", user.LoginHandler())
			userGroup.PUT("/update", middleware.Auth(true), user.UpdateHandler())
		}

		postGroup := r.Group("/post", middleware.Auth(true))
		{
			postGroup.POST("/publish_confession", confession.PublishConfessionHandler())
			postGroup.POST("/update_confession", confession.UpdateConfessionHandler())
			postGroup.GET("/all_list", confession.GetConfessionHandler())
			postGroup.GET("/my_list", confession.GetMyConfessionsHandler())
			postGroup.DELETE("/delete_post", confession.DeleteConfessionHandler())
		}

		blockGroup := r.Group("/block", middleware.Auth(true))
		{
			blockGroup.POST("/block", middleware.Auth(true), block.BlockHandler())
			blockGroup.POST("/unblock", middleware.Auth(true), block.UnblockHandler())
			blockGroup.POST("/list_block", middleware.Auth(true), block.GetBlockedHandler())
		}
	}
}

func routePrefix() string {
	return "/api"
}

func routeBase(r *gin.RouterGroup, router *gin.Engine) {
	// OpenAPI/Swagger 文档生成
	if slices.Contains([]string{config.AppEnvDev, config.AppEnvTest}, config.AppEnv()) {
		r.GET("/swagger.json", swagger.DocumentHandler(router))
	}

	// 健康检查
	r.GET("/health", api.HealthHandler())
}
