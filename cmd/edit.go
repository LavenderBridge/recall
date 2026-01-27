package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/LavenderBridge/spaced-repetition/internal/models"
	"github.com/spf13/cobra"
)

var (
	editName       string
	editURL        string
	editNotes      string
	editTags       string
	editDifficulty int
)

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a problem details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("❌ Invalid ID")
			return
		}

		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		// Get existing
		// We need a GetProblemByID, but currently we only have GetProblemByName.
		// Let's rely on List and filtering for now or add GetProblemByID.
		// Actually, let's implement GetProblemByID nicely or just List and find.
		// Listing all is inefficient but fine for CLI MVP. 
		// BETTER: Implementing GetProblemByID in DB is cleaner. 
		// Wait, I can't restart DB implementation easily. 
		// Let's use ListProblems and find the ID.
		
		problems, err := store.ListProblems(false)
		if err != nil {
			fmt.Println("❌ Error fetching problems:", err)
			return
		}
		
		var target *models.Problem
		for _, p := range problems {
			if p.ID == id {
				target = &p
				break
			}
		}

		if target == nil {
			fmt.Println("❌ Problem not found with ID:", id)
			return
		}

		// Apply updates
		if cmd.Flags().Changed("name") {
			target.Name = editName
		}
		if cmd.Flags().Changed("url") {
			target.URL = editURL
		}
		if cmd.Flags().Changed("notes") {
			target.Notes = editNotes
		}
		if cmd.Flags().Changed("difficulty") {
			if editDifficulty < 1 || editDifficulty > 5 {
				fmt.Println("❌ Difficulty must be between 1 and 5")
				return
			}
			target.Difficulty = editDifficulty
		}
		if cmd.Flags().Changed("tags") {
			var newTags []models.Tag
			parts := strings.Split(editTags, ",")
			for _, part := range parts {
				t := strings.TrimSpace(part)
				if t != "" {
					newTags = append(newTags, models.Tag{Name: t})
				}
			}
			target.Tags = newTags
		}

		// Save
		if err := store.UpdateProblemDetails(*target); err != nil {
			fmt.Println("❌ Error updating problem:", err)
			return
		}

		fmt.Println("✅ Problem updated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVar(&editName, "name", "", "New name")
	editCmd.Flags().StringVar(&editURL, "url", "", "New URL")
	editCmd.Flags().StringVar(&editNotes, "notes", "", "New notes")
	editCmd.Flags().IntVar(&editDifficulty, "difficulty", 0, "New difficulty (1-5)")
	editCmd.Flags().StringVar(&editTags, "tags", "", "Comma-separated tags (replaces existing)")
}
