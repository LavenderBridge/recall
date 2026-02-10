package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LavenderBridge/spaced-repetition/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := filepath.Join(home, ".recall")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create data directory: %w", err)
	}

	dbPath := filepath.Join(dir, "recall.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func initSchema(db *sql.DB) error {
	// Problems table
	// We use IF NOT EXISTS. For migration, we might need manual ALTER if columns missing.
	// For this MVP refactor, we'll just try to add columns if they don't exist logic is a bit complex for simple SQL.
	// We'll define the ideal schema. If table exists but old schema, it might fail.
	// Let's rely on adding columns gracefully or just CREATE IF NOT EXISTS for new tables.
	// Since we are refactoring, let's assume we can try to ALTER TABLE for new columns.

	query := `
	CREATE TABLE IF NOT EXISTS problems (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		url TEXT,
		notes TEXT,
		difficulty INTEGER NOT NULL,
		interval INTEGER DEFAULT 1,
		ease_factor REAL DEFAULT 2.5,
		last_reviewed DATE NOT NULL,
		next_review DATE NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Tags table
	queryTags := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);
	`
	if _, err := db.Exec(queryTags); err != nil {
		return err
	}

	// Join table
	queryProblemTags := `
	CREATE TABLE IF NOT EXISTS problem_tags (
		problem_id INTEGER,
		tag_id INTEGER,
		PRIMARY KEY (problem_id, tag_id),
		FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(queryProblemTags); err != nil {
		return err
	}

	// Migrations (Simple check and Apply)
	// Check if 'url' exists in 'problems'
	if !columnExists(db, "problems", "url") {
		db.Exec("ALTER TABLE problems ADD COLUMN url TEXT")
		db.Exec("ALTER TABLE problems ADD COLUMN notes TEXT")
		db.Exec("ALTER TABLE problems ADD COLUMN interval INTEGER DEFAULT 1")
		db.Exec("ALTER TABLE problems ADD COLUMN ease_factor REAL DEFAULT 2.5")
	}

	// Reviews table
	queryReviews := `
	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		problem_id INTEGER,
		quality INTEGER,
		reviewed_at DATETIME,
		notes TEXT,
		interval_snapshot INTEGER,
		ease_factor_snapshot REAL,
		FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(queryReviews); err != nil {
		return err
	}

	if err := migrateLegacyReviews(db); err != nil {
		fmt.Printf("⚠️ Migration warning: %v\n", err)
	}

	return nil
}

func columnExists(db *sql.DB, tableName, colName string) bool {
	// Simple check by selecting 1 limit 0
	_, err := db.Query(fmt.Sprintf("SELECT %s FROM %s LIMIT 0", colName, tableName))
	return err == nil
}

func (s *Store) AddProblem(p models.Problem) error {
	res, err := s.db.Exec(`
		INSERT INTO problems (name, url, notes, difficulty, interval, ease_factor, last_reviewed, next_review)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.URL, p.Notes, p.Difficulty, p.Interval, p.EaseFactor, p.LastReviewed, p.NextReview,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Add tags
	for _, tag := range p.Tags {
		if err := s.linkTag(int(id), tag.Name); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) linkTag(problemID int, tagName string) error {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return nil
	}

	// Ensure tag exists
	_, err := s.db.Exec(`INSERT OR IGNORE INTO tags (name) VALUES (?)`, tagName)
	if err != nil {
		return err
	}

	// Get Tag ID
	var tagID int
	err = s.db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err != nil {
		return err
	}

	// Link
	_, err = s.db.Exec(`INSERT OR IGNORE INTO problem_tags (problem_id, tag_id) VALUES (?, ?)`, problemID, tagID)
	return err
}

func (s *Store) GetProblem(name string) (*models.Problem, error) {
	row := s.db.QueryRow(`
		SELECT id, name, url, notes, difficulty, interval, ease_factor, last_reviewed, next_review
		FROM problems WHERE name = ?`, name)
	
	var p models.Problem
	var url, notes sql.NullString
	err := row.Scan(&p.ID, &p.Name, &url, &notes, &p.Difficulty, &p.Interval, &p.EaseFactor, &p.LastReviewed, &p.NextReview)
	if err != nil {
		return nil, err
	}
	p.URL = url.String
	p.Notes = notes.String
	
	p.Tags, _ = s.getTagsForProblem(p.ID)
	return &p, nil
}

func (s *Store) UpdateProblem(p models.Problem) error {
	_, err := s.db.Exec(`
		UPDATE problems
		SET difficulty=?, interval=?, ease_factor=?, last_reviewed=?, next_review=?
		WHERE id=?`,
		p.Difficulty, p.Interval, p.EaseFactor, p.LastReviewed, p.NextReview, p.ID,
	)
	return err
}

func (s *Store) UpdateProblemDetails(p models.Problem) error {
	// 1. Update core fields
	query := `
		UPDATE problems
		SET name=?, url=?, notes=?, difficulty=?
		WHERE id=?
	`
	_, err := s.db.Exec(query, p.Name, p.URL, p.Notes, p.Difficulty, p.ID)
	if err != nil {
		return err
	}

	// 2. Update tags (Full replace strategy: Delete all then re-add works, or diff. 
	//    For MVP CLI, clear and re-add is simplest safe approach)
	
	// Only update tags if the list is not nil (empty list might mean remove all, nil means don't touch)
	// But in our struct Tag is a slice, so it's either empty or has items. 
	// We need a way to know if we should update tags.
	// For now, let's assume if this is called, we want to set the tags to whatever is in p.Tags
	// Note: careful if we just want to update name and keep tags. 
	// The caller should ideally provide the full object or we need a way to partial update.
	// Given the CLI 'edit' command will probably read existing, apply flags, then save, 
	// we can assume p.Tags contains the final desired state.

	_, err = s.db.Exec("DELETE FROM problem_tags WHERE problem_id=?", p.ID)
	if err != nil {
		return err
	}

	for _, tag := range p.Tags {
		if err := s.linkTag(p.ID, tag.Name); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) DeleteProblem(id int) error {
	// Cascading delete should handle problem_tags if defined in schema, 
	// but let's be explicit if schema didn't have cascade (which we did put in initSchema).
	_, err := s.db.Exec("DELETE FROM problems WHERE id=?", id)
	return err
}

func (s *Store) ListProblems(dueOnly bool) ([]models.Problem, error) {
	var query string
	if dueOnly {
		query = `SELECT id, name, url, notes, difficulty, interval, ease_factor, last_reviewed, next_review FROM problems WHERE date(next_review) <= date('now') ORDER BY next_review ASC`
	} else {
		query = `SELECT id, name, url, notes, difficulty, interval, ease_factor, last_reviewed, next_review FROM problems ORDER BY next_review ASC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		var url, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &url, &notes, &p.Difficulty, &p.Interval, &p.EaseFactor, &p.LastReviewed, &p.NextReview); err != nil {
			return nil, err
		}
		p.URL = url.String
		p.Notes = notes.String
		// Optimization: Could fetch tags in batch, but for CLI loop is okay
		p.Tags, _ = s.getTagsForProblem(p.ID)
		problems = append(problems, p)
	}
	return problems, nil
}

func (s *Store) getTagsForProblem(problemID int) ([]models.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name 
		FROM tags t
		JOIN problem_tags pt ON t.id = pt.tag_id
		WHERE pt.problem_id = ?`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		rows.Scan(&t.ID, &t.Name)
		tags = append(tags, t)
	}
	return tags, nil
}

// migrateLegacyReviews backfills the reviews table from problems that have been reviewed but have no history.
func migrateLegacyReviews(db *sql.DB) error {
	// Find problems that have been reviewed (last_reviewed > epoch)
	rows, err := db.Query(`SELECT id, last_reviewed, interval, ease_factor FROM problems WHERE last_reviewed IS NOT NULL And last_reviewed > '1970-01-01'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type legacyProb struct {
		ID           int
		LastReviewed time.Time
		Interval     int
		EaseFactor   float64
	}

	var toMigrate []legacyProb

	for rows.Next() {
		var p legacyProb
		if err := rows.Scan(&p.ID, &p.LastReviewed, &p.Interval, &p.EaseFactor); err != nil {
			continue
		}
		toMigrate = append(toMigrate, p)
	}

	for _, p := range toMigrate {
		// Check if any review exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM reviews WHERE problem_id = ?", p.ID).Scan(&count)
		if err != nil {
			continue
		}

		if count == 0 {
			// Insert synthetic review
			_, err := db.Exec(`
				INSERT INTO reviews (problem_id, quality, reviewed_at, notes, interval_snapshot, ease_factor_snapshot)
				VALUES (?, ?, ?, ?, ?, ?)`,
				p.ID, 3, p.LastReviewed, "Imported from legacy data", p.Interval, p.EaseFactor,
			)
			if err != nil {
				fmt.Printf("Failed to migrate problem %d: %v\n", p.ID, err)
			}
		}
	}
	return nil
}

func (s *Store) AddReview(r models.Review) error {
	_, err := s.db.Exec(`
		INSERT INTO reviews (problem_id, quality, reviewed_at, notes, interval_snapshot, ease_factor_snapshot)
		VALUES (?, ?, ?, ?, ?, ?)`,
		r.ProblemID, r.Quality, r.ReviewedAt, r.Notes, r.Interval, r.EaseFactor,
	)
	return err
}

func (s *Store) GetLastReview(problemID int) (*models.Review, error) {
	row := s.db.QueryRow(`
		SELECT id, problem_id, quality, reviewed_at, notes, interval_snapshot, ease_factor_snapshot
		FROM reviews 
		WHERE problem_id = ? 
		ORDER BY reviewed_at DESC 
		LIMIT 1`, problemID)

	var r models.Review
	var notes sql.NullString
	if err := row.Scan(&r.ID, &r.ProblemID, &r.Quality, &r.ReviewedAt, &notes, &r.Interval, &r.EaseFactor); err != nil {
		return nil, err
	}
	r.Notes = notes.String
	return &r, nil
}

func (s *Store) GetReviewStats() (*models.ReviewStats, error) {
	stats := &models.ReviewStats{
		CountByDifficulty: make(map[int]int),
	}

	// Total Reviews
	if err := s.db.QueryRow("SELECT COUNT(*) FROM reviews").Scan(&stats.TotalReviews); err != nil {
		return nil, err
	}

	// Reviews Last 7 Days
	if err := s.db.QueryRow("SELECT COUNT(*) FROM reviews WHERE reviewed_at > date('now', '-7 days')").Scan(&stats.ReviewsLast7Days); err != nil {
		return nil, err
	}

	// Average Quality
	var avg sql.NullFloat64
	if err := s.db.QueryRow("SELECT AVG(quality) FROM reviews").Scan(&avg); err != nil {
		return nil, err
	}
	if avg.Valid {
		stats.AverageQuality = avg.Float64
	}

	// Breakdown by difficulty (from problems table, not reviews, usually)
	rows, err := s.db.Query("SELECT difficulty, COUNT(*) FROM problems GROUP BY difficulty")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var diff, count int
		if err := rows.Scan(&diff, &count); err == nil {
			stats.CountByDifficulty[diff] = count
		}
	}

	return stats, nil
}
