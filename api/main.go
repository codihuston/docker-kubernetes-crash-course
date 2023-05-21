package main

import (
	"net/http"

	controllers "example.com/m/v2/controllers"
	db "example.com/m/v2/db"
	ioc "example.com/m/v2/ioc"
	middleware "example.com/m/v2/middleware"
	repositories "example.com/m/v2/repositories"
	routers "example.com/m/v2/routers"
	services "example.com/m/v2/services"
	"github.com/gin-gonic/gin"
)

func main() {
	ioc := ioc.NewContainer()
	db := db.Connect()
	r := gin.Default()

	// Init error handler
	r.Use(middleware.ErrorHandler(&ioc))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	blogController := controllers.NewBlogController(
		&ioc,
		services.NewBlogService(
			&ioc,
		),
		repositories.NewPostgreSQLBlogRepository(
			&ioc,
			db,
		),
	)

	routers.InitBlogRouter(r, blogController)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
