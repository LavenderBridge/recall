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

// Review represents a single review event for a problem.
type Review struct {
	ID         int       `json:"id"`
	ProblemID  int       `json:"problem_id"`
	Quality    int       `json:"quality"`
	ReviewedAt time.Time `json:"reviewed_at"`
	Notes      string    `json:"notes"`
	// Snapshot of algorithm state at time of review
	Interval   int     `json:"interval"`
	EaseFactor float64 `json:"ease_factor"`
}

type ReviewStats struct {
	TotalReviews      int
	ReviewsLast7Days  int
	AverageQuality    float64
	CountByDifficulty map[int]int
}
