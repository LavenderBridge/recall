package models

import "time"

// Problem represents a single practice item.
type Problem struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Notes        string    `json:"notes"`
	Difficulty   int       `json:"difficulty"`    // Initial difficulty (1-5)
	Interval     int       `json:"interval"`      // Days until next review
	EaseFactor   float64   `json:"ease_factor"`   // SM-2 multiplier
	LastReviewed time.Time `json:"last_reviewed"`
	NextReview   time.Time `json:"next_review"`
	Tags         []Tag     `json:"tags,omitempty"`
}

// Tag represents a category for a problem.
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
