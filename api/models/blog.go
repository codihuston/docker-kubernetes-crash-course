package models

import (
	"strings"

	"gorm.io/gorm"

	utils "example.com/m/v2/pkg/utils"
)

type Blog struct {
	gorm.Model
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

func (b *Blog) GetWordCount() map[string]int {
	m := make(map[string]int)

	words := strings.Split(b.Body, " ")

	// For each word
	for _, word := range words {
		// Check if in map
		_, ok := m[utils.ReplaceSymbols(word)]
		if ok {
			// If so, increment
			m[word] += 1
		} else {
			// Otherwise, init to 1
			m[word] = 1
		}
	}

	return m
}
