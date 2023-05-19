package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	models "example.com/m/v2/models"
	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRESQL_URL")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/blogs", func(c *gin.Context) {
		var blogs []models.Blog
		db.Find(&blogs)
		c.JSON(http.StatusOK, blogs)
	})
	r.GET("/blogs/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")

		var blog models.Blog
		db.Find(&blog, id)
		c.JSON(http.StatusOK, blog)
	})
	r.GET("/blogs/:id/words", func(c *gin.Context) {
		id := c.Params.ByName("id")

		var blog models.Blog
		db.Find(&blog, id)
		wc := blog.GetWordCount()

		c.JSON(http.StatusOK, wc)
	})
	r.POST("/blogs", func(c *gin.Context) {
		var blog models.Blog
		c.BindJSON(&blog)
		db.Create(&models.Blog{Title: blog.Title, Body: blog.Body})
	})
	r.PUT("/blogs/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 64)
		if err != nil {
			fmt.Println(err)
		}

		var blog models.Blog
		c.BindJSON(&blog)

		blog.ID = uint(id)

		db.Save(&blog)
		c.JSON(http.StatusOK, blog)
	})
	r.DELETE("/blogs/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		db.Delete(&models.Blog{}, id)
		c.JSON(http.StatusOK, nil)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
