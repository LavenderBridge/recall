package algorithm

import (
	"math"
	"time"

	"github.com/LavenderBridge/spaced-repetition/internal/models"
)

// Default settings for new problems
const (
	InitialInterval   = 1
	InitialEaseFactor = 2.5
)

// CalculateReview updates the problem based on the quality of the review session.
// quality: 0 (blackout) to 5 (perfect recollection)
func CalculateReview(p models.Problem, quality int) models.Problem {
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}

	// 1. Calculate new Ease Factor
	// EF' = EF + (0.1 - (5-q) * (0.08 + (5-q)*0.02))
	// If EF goes below 1.3, set it to 1.3
	q := float64(quality)
	newEase := p.EaseFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if newEase < 1.3 {
		newEase = 1.3
	}

	// 2. Calculate new Interval
	var newInterval int
	if quality < 3 {
		// If the user failed (quality < 3), start over
		newInterval = 1
		// Optionally, we could reset interval to 1 but keep EF or reduce it.
		// Standard SM-2 resets interval to 1.
	} else {
		if p.Interval == 0 {
			newInterval = 1
		} else if p.Interval == 1 {
			newInterval = 6
		} else {
			newInterval = int(math.Ceil(float64(p.Interval) * newEase))
		}
	}

	// Update the problem struct
	p.EaseFactor = newEase
	p.Interval = newInterval
	p.LastReviewed = time.Now()
	p.NextReview = p.LastReviewed.AddDate(0, 0, newInterval)

	return p
}

// InitProblem sets default values for a new problem
func InitProblem(p models.Problem, initialEase float64) models.Problem {
	if initialEase == 0 {
		initialEase = InitialEaseFactor
	}
	p.EaseFactor = initialEase
	p.Interval = InitialInterval
	p.LastReviewed = time.Now()
	// Next review is initially tomorrow? Or today if we assume valid immediately.
	// Typically immediate review -> then interval. But mostly for CLI we set it for 1 day later or same day.
	// Let's set it to 1 day later for now as "Learning" phase usually implies reviewing next day.
	p.NextReview = p.LastReviewed.AddDate(0, 0, InitialInterval)
	return p
}
