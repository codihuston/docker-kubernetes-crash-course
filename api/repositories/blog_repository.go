package repositories

import (
	"example.com/m/v2/ioc"
	models "example.com/m/v2/models"
	"gorm.io/gorm"
)

// Note: the use of these smaller interfaces enables our mock repository can
// be smaller, enabling us to add functionality as needed. This way, our mock
// code can remain as small as possible. If we used a single interface,
// the db instance passed into the Service's methods must conform to 100%
// of the defined interface, when in reality, we only care about a small part
// of that interface at one given time.
type BlogCreator interface {
	Create(m *models.Blog) (*models.Blog, error)
}

type MultiBlogGetter interface {
	GetAll() ([]*models.Blog, error)
}

type SingleBlogGetter interface {
	GetByID(id string) (*models.Blog, error)
}

type BlogUpdater interface {
	Update(id uint, m *models.Blog) (*models.Blog, error)
}

type BlogDeleter interface {
	Delete(id string) error
}

type BlogRepository interface {
	BlogCreator
	MultiBlogGetter
	SingleBlogGetter
	BlogUpdater
	BlogDeleter
}

type PostgreSQLBlogRepository struct {
	ioc *ioc.IOC
	db  *gorm.DB
}

// TODO: so the article says returning BlogRepository here would be bad
// b/c it'd force the tester to have to mock everything!
func NewPostgreSQLBlogRepository(c *ioc.IOC, db *gorm.DB) *PostgreSQLBlogRepository {
	return &PostgreSQLBlogRepository{
		ioc: c,
		db:  db,
	}
}

func (r *PostgreSQLBlogRepository) Create(m *models.Blog) (*models.Blog, error) {
	if err := r.db.Save(&m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *PostgreSQLBlogRepository) GetByID(id string) (*models.Blog, error) {
	var m models.Blog
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostgreSQLBlogRepository) GetAll() ([]*models.Blog, error) {
	var m []*models.Blog
	if err := r.db.Find(&m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *PostgreSQLBlogRepository) Update(id uint, m *models.Blog) (*models.Blog, error) {
	// TODO: make immutable by fetching, merging, and persisting
	m.ID = id
	if err := r.db.Save(&m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (r *PostgreSQLBlogRepository) Delete(id string) error {
	if err := r.db.Delete(&models.Blog{}, id).Error; err != nil {
		return err
	}
	return nil
}
