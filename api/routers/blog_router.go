package routers

import (
	"example.com/m/v2/controllers"
	"github.com/gin-gonic/gin"
)

// Note: can use interface controllers.Controller if the default interface
// suits you. BlogController uses a non-standard interface for ShowWordCount().
func InitBlogRouter(r *gin.Engine, controller controllers.BlogController) *gin.RouterGroup {
	routes := r.Group("/blogs")
	{
		// NOTE: gin requires trailing slash!
		routes.GET("/", controller.Index)
		routes.GET("/new", controller.New)
		routes.GET("/:id", controller.Show)
		routes.GET("/:id/words", controller.ShowWordCount)
		routes.POST("/", controller.Create)
		routes.PUT("/:id", controller.Update)
		routes.DELETE("/:id", controller.Delete)
	}
	return routes
}
