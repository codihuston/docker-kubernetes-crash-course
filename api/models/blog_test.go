package models

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordCount(t *testing.T) {
	var blog = &Blog{
		Title: "test title",
		Body:  "red red red blue green green yellow yellow yellow yellow",
	}

	result := blog.GetWordCount()
	expected := map[string]int{
		"blue":   1,
		"green":  2,
		"red":    3,
		"yellow": 4,
	}

	assert.True(t, reflect.DeepEqual(result, expected), "The two word counts be the same.")
}
