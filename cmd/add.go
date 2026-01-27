package cmd

import (
	"fmt"
	"strings"

	"github.com/LavenderBridge/spaced-repetition/internal/algorithm"
	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/LavenderBridge/spaced-repetition/internal/models"
	"github.com/spf13/cobra"
)

var (
	addURL   string
	addNotes string
	addTags  string
)

var addCmd = &cobra.Command{
	Use:   "add [name] [difficulty 1-5]",
	Short: "Add a new problem to track",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		diffStr := args[1]

		difficulty := 0
		fmt.Sscanf(diffStr, "%d", &difficulty)
		if difficulty < 1 || difficulty > 5 {
			fmt.Println("❌ Difficulty must be between 1 and 5")
			return
		}

		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		// Parse tags
		var tags []models.Tag
		if addTags != "" {
			parts := strings.Split(addTags, ",")
			for _, part := range parts {
				tags = append(tags, models.Tag{Name: strings.TrimSpace(part)})
			}
		}

		problem := models.Problem{
			Name:       name,
			URL:        addURL,
			Notes:      addNotes,
			Difficulty: difficulty,
			Tags:       tags,
		}

		// Initialize SM-2 values
		problem = algorithm.InitProblem(problem, 0)

		if err := store.AddProblem(problem); err != nil {
			fmt.Println("❌ Error adding problem:", err)
			return
		}

		fmt.Printf("✅ Added '%s' (Next review: %s)\n", name, problem.NextReview.Format("2006-01-02"))
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addURL, "url", "u", "", "URL to the problem")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Notes about the problem")
	addCmd.Flags().StringVarP(&addTags, "tags", "t", "", "Comma-separated tags (e.g. array,dp)")
}
