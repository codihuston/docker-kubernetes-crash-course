package controllers

import (
	"github.com/gin-gonic/gin"

	apiErrors "example.com/m/v2/errors"
)

// NOTE: this is based off of Ruby on Rails' CRUD verbs.
// See: https://guides.rubyonrails.org/routing.html#crud-verbs-and-actions
type Controller interface {
	Index(*gin.Context)  // Display a list of all resources
	Show(*gin.Context)   // Display a specific resource
	New(*gin.Context)    // Return a form for creating a new resource
	Create(*gin.Context) // Create a new resource
	Edit(*gin.Context)   // Return a form for editing a resource
	Update(*gin.Context) // Update a specific resource
	Delete(*gin.Context) // Delete a specific resource
}

// HandleAPIError will take a plan error object and parse it into into
// an APIError. If the type of error is not handled, it will become
// a simple error.
func HandleAPIError(c *gin.Context, err error) {
	// Process the error for known errors mapped to specific HTTP responses and
	// messages
	apiError := apiErrors.NewAPIError(err)

	// Set the response body for the client.
	// NOTE: Gin will not return a body for `no content` status codes,
	// such as 204.
	c.JSON(apiError.Code, gin.H{
		"code":    apiError.Code,
		"message": apiError.GetMessage(),
	})

	// Invoke error handler with error code for client and a message for
	// the server. Do not override the error code.
	c.AbortWithError(-1, apiError.Unwrap())
	return
}
