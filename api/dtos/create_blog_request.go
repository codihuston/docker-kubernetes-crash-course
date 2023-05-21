package dtos

// This struct decouples the HTTP payload from the internal model structure.
// This feels like duplication, but is intentional.
type CreateBlogRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type UpdateBlogRequest struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}
