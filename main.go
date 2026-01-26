package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "leetcode.db"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	db := initDB()
	defer db.Close()

	switch os.Args[1] {
	case "add":
		handleAdd(db)
	case "due":
		handleDue(db)
	case "update":
		handleUpdate(db)
	case "list":
		handleList(db)
	default:
		printUsage()
	}
}

/* -------------------- COMMAND HANDLERS -------------------- */

func handleAdd(db *sql.DB) {
	if len(os.Args) != 4 {
		fmt.Println("Usage: leetcode add \"Problem Name\" <difficulty 1-5>")
		return
	}

	name := os.Args[2]
	diff := mustDifficulty(os.Args[3])

	now := time.Now()
	next := nextReview(diff, now)

	_, err := db.Exec(`
		INSERT INTO problems (name, difficulty, last_reviewed, next_review)
		VALUES (?, ?, ?, ?)`,
		name, diff, now, next,
	)

	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("✅ Added:", name)
}

func handleDue(db *sql.DB) {
	rows, err := db.Query(`
		SELECT name, difficulty, next_review
		FROM problems
		WHERE date(next_review) <= date('now')
		ORDER BY next_review ASC
	`)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer rows.Close()

	found := false
	fmt.Println("🔥 Problems due today:\n")

	for rows.Next() {
		found = true
		var name string
		var diff int
		var next string

		rows.Scan(&name, &diff, &next)
		fmt.Printf("- %s (last diff: %d)\n", name, diff)
	}

	if !found {
		fmt.Println("✅ No problems due today. Brain = fresh.")
	}
}

func handleUpdate(db *sql.DB) {
	if len(os.Args) != 4 {
		fmt.Println("Usage: leetcode update \"Problem Name\" <difficulty 1-5>")
		return
	}

	name := os.Args[2]
	diff := mustDifficulty(os.Args[3])

	now := time.Now()
	next := nextReview(diff, now)

	res, err := db.Exec(`
		UPDATE problems
		SET difficulty = ?, last_reviewed = ?, next_review = ?
		WHERE name = ?`,
		diff, now, next, name,
	)

	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		fmt.Println("⚠️ Problem not found")
		return
	}

	fmt.Println("🔄 Updated:", name)
}

func handleList(db *sql.DB) {
	rows, err := db.Query(`
		SELECT name, difficulty, last_reviewed, next_review
		FROM problems
		ORDER BY next_review ASC
	`)
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("📚 All tracked problems:\n")

	for rows.Next() {
		var name string
		var diff int
		var last, next string

		rows.Scan(&name, &diff, &last, &next)
		fmt.Printf("- %s | diff:%d | next:%s\n", name, diff, next)
	}
}

/* -------------------- CORE LOGIC -------------------- */

func nextReview(difficulty int, from time.Time) time.Time {
	intervals := map[int]int{
		1: 14,
		2: 7,
		3: 4,
		4: 2,
		5: 1,
	}
	return from.AddDate(0, 0, intervals[difficulty])
}

/* -------------------- DB INIT -------------------- */

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS problems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			difficulty INTEGER NOT NULL,
			last_reviewed DATE NOT NULL,
			next_review DATE NOT NULL
		)
	`)
	if err != nil {
		panic(err)
	}

	return db
}

/* -------------------- HELPERS -------------------- */

func mustDifficulty(arg string) int {
	diff, err := strconv.Atoi(arg)
	if err != nil || diff < 1 || diff > 5 {
		fmt.Println("❌ Difficulty must be between 1 and 5")
		os.Exit(1)
	}
	return diff
}

func printUsage() {
	fmt.Println(`
LeetCode Spaced Repetition CLI

Commands:
  add "Problem Name" <1-5>     Add new problem
  update "Problem Name" <1-5>  Update difficulty after recap
  due                          Show problems due today
  list                         List all problems
`)
}

