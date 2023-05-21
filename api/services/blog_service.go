package services

import (
	dtos "example.com/m/v2/dtos"
	"example.com/m/v2/ioc"
	models "example.com/m/v2/models"
	repositories "example.com/m/v2/repositories"
)

// BlogService handles business logic related to blogs
type blogService struct {
	ioc *ioc.IOC
}

type BlogService interface {
	Create(m *dtos.CreateBlogRequest, r repositories.BlogCreator) (*models.Blog, error)
	GetByID(id string, r repositories.SingleBlogGetter) (*models.Blog, error)
	GetAll(r repositories.MultiBlogGetter) ([]*models.Blog, error)
	Update(id uint, m *dtos.UpdateBlogRequest, r repositories.BlogUpdater) (*models.Blog, error)
	Delete(id string, r repositories.BlogDeleter) error
}

func NewBlogService(c *ioc.IOC) *blogService {
	return &blogService{
		ioc: c,
	}
}

// Note: the use of the smaller repository interfaces allow us to have slimmer
// mocks.
func (s blogService) Create(m *dtos.CreateBlogRequest, r repositories.BlogCreator) (*models.Blog, error) {
	res, err := r.Create(s.mapCreateBlogRequestToModel(*m))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s blogService) GetByID(id string, r repositories.SingleBlogGetter) (*models.Blog, error) {
	res, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s blogService) GetAll(r repositories.MultiBlogGetter) ([]*models.Blog, error) {
	res, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: should id's be string or uint? Make consistent everywhere else!
func (s blogService) Update(id uint, m *dtos.UpdateBlogRequest, r repositories.BlogUpdater) (*models.Blog, error) {
	res, err := r.Update(id, s.mapUpdateBlogRequestToModel(*m))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s blogService) Delete(id string, r repositories.BlogDeleter) error {
	err := r.Delete(id)
	return err
}

func (s blogService) mapCreateBlogRequestToModel(request dtos.CreateBlogRequest) *models.Blog {
	// Perform mapping or conversion from DTO to domain model
	model := &models.Blog{
		Title: request.Title,
		Body:  request.Body,
	}
	return model
}

func (s blogService) mapUpdateBlogRequestToModel(request dtos.UpdateBlogRequest) *models.Blog {
	// Perform mapping or conversion from DTO to domain model
	model := &models.Blog{
		Title: request.Title,
		Body:  request.Body,
	}
	return model
}
