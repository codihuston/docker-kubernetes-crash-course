package controllers

import (
	"net/http"
	"strconv"

	dtos "example.com/m/v2/dtos"
	apiErrors "example.com/m/v2/errors"
	"example.com/m/v2/ioc"
	"example.com/m/v2/repositories"
	services "example.com/m/v2/services"

	"github.com/gin-gonic/gin"
)

type blogController struct {
	blogService    services.BlogService
	ioc            *ioc.IOC
	blogRepository repositories.BlogRepository
}

type BlogController interface {
	Controller
	ShowWordCount(*gin.Context)
}

func NewBlogController(c *ioc.IOC, s services.BlogService, r repositories.BlogRepository) *blogController {
	return &blogController{
		ioc:            c,
		blogService:    s,
		blogRepository: r,
	}
}

func (b *blogController) Index(c *gin.Context) {
	res, err := b.blogService.GetAll(b.blogRepository)

	if err != nil {
		HandleAPIError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (b *blogController) Show(c *gin.Context) {
	id := c.Params.ByName("id")
	if _, err := strconv.Atoi(id); err != nil {
		// fmt.Printf("%q does not look like a number.\n", v)
		c.Next()
		return
	}

	res, err := b.blogService.GetByID(id, b.blogRepository)
	if err != nil {
		HandleAPIError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (b *blogController) ShowWordCount(c *gin.Context) {
	id := c.Params.ByName("id")

	res, err := b.blogService.GetByID(id, b.blogRepository)
	if err != nil {
		HandleAPIError(c, err)
		return
	}

	wc := res.GetWordCount()
	c.JSON(http.StatusOK, wc)
}

func (b *blogController) New(c *gin.Context) {
	// Note: this code is reachable by clients. We might test for such
	// specific behavior as an integration or E2E test.
	err := apiErrors.IsNotImplementedError
	HandleAPIError(c, err)
}

func (b *blogController) Create(c *gin.Context) {
	reqBody := &dtos.CreateBlogRequest{}
	c.BindJSON(reqBody)
	res, err := b.blogService.Create(reqBody, b.blogRepository)

	if err != nil {
		HandleAPIError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (b *blogController) Edit(c *gin.Context) {
	// Note: this code is not reachable by clients. We might test for such
	// specific behavior as an integration or E2E test.
	err := apiErrors.IsNotImplementedError
	HandleAPIError(c, err)
}

func (b *blogController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 64)

	if err != nil {
		HandleAPIError(c, err)
		return
	}

	reqBody := &dtos.UpdateBlogRequest{}
	c.BindJSON(reqBody)
	res, err := b.blogService.Update(uint(id), reqBody, b.blogRepository)

	if err != nil {
		HandleAPIError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (b *blogController) Delete(c *gin.Context) {
	// TODO: parse uint?
	id := c.Params.ByName("id")
	err := b.blogService.Delete(id, b.blogRepository)

	if err != nil {
		HandleAPIError(c, err)
		return
	}

	c.Writer.WriteHeader(204)
}
