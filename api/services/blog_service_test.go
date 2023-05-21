package services

import (
	"errors"
	"testing"

	dtos "example.com/m/v2/dtos"
	"example.com/m/v2/ioc"
	models "example.com/m/v2/models"
	mocks "example.com/m/v2/repositories/mocks"
)

func TestCreate(t *testing.T) {
	c := ioc.NewContainer()
	s := NewBlogService(&c)
	d := &dtos.CreateBlogRequest{
		Title: "my first blog post",
		Body:  "hello world!",
	}

	// Define a list of individual test scenarios, their mocks, and
	// whether or not they should error out. If there is no mock method defined,
	// the default method in the mock is used instead. This pattern allows us to
	// mock using golang out-of-box, without any 3rd party dependencies
	// (testify/mock). Though, you could probably even extend this pattern's
	// capabilities by using such dependencies.
	tests := [...]struct {
		name      string
		store     *mocks.BlogRepositoryMock
		shouldErr bool
	}{
		{
			"HappyPath",
			nil,
			false,
		},
		{
			"Negative",
			&mocks.BlogRepositoryMock{
				MockCreate: func(m *models.Blog) (*models.Blog, error) {
					return nil, errors.New("generic error")
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Create(d, tt.store)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got <nil>")
			}
		})
	}
}
