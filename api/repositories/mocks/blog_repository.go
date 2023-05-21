package mocks

import (
	models "example.com/m/v2/models"
)

// IMPORTANT: this struct must match function signature of the Repository being
// used in the corresponding Service method. We do that with the pointer
// receiver functions explicitly defined below.
//
// The `Mock*` field functions that
// that are declared but not implemented are intended to allow a tester to
// override mock behaviour ad hoc. The use of public members as an option to
// override mock functionality is a requirement when the separation of mocks
// and services on the file system is preferred.
//
// The term "override" is used loosely here, as it is not true overriding like
// in OOP.
//
// This file contains the default functionality of each mocked method.
type BlogRepositoryMock struct {
	MockCreate func(m *models.Blog) (*models.Blog, error)
}

// Note: so long as we handle the nil case of `mock`, we are allowed to do the
// following check, which allows use to override this mock function at will
// outside of this package.
//
// Signature           : func (mock *blogRepositoryMock) Create(m *models.Blog) (*models.Blog, error)
// Is the exact same as: func Create(mock *blogRepositoryMock, m *models.Blog) (*models.Blog, error)
func (mock *BlogRepositoryMock) Create(m *models.Blog) (*models.Blog, error) {
	if mock != nil && mock.MockCreate != nil {
		return mock.MockCreate(m)
	}

	return &models.Blog{
		// ID:        1,
		Title: "my first blog post",
		Body:  "hello world!",
		// CreatedAt: "2023-05-24T08:19:50.99933Z",
		// UpdatedAt: "2023-05-24T08:19:50.99933Z",
	}, nil
}
