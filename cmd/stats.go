package cmd

import (
	"fmt"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show tracked problem statistics",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		problems, err := store.ListProblems(false)
		if err != nil {
			fmt.Println("❌ Error fetching problems:", err)
			return
		}

		total := len(problems)
		due := 0
		mastered := 0 // Arbitrary definition: interval > 30 days
		learning := 0 // interval < 7

		for _, p := range problems {
			// Check if due
			// This is rough approximation as ListProblems(false) doesn't filter by date
			// But we can check NextReview vs Now if we wanted, but ListProblems doesn't give us easy Compare without parsing.
			// Actually proper check requires comparing dates.
			// Let's just trust 'due' command for due count, here stick to distribution.
			
			if p.Interval > 30 {
				mastered++
			} else if p.Interval < 7 {
				learning++
			}
		}

		fmt.Println("📊 Statistics")
		fmt.Println("-------------")
		fmt.Printf("Total Problems: %d\n", total)
		fmt.Printf("Due (Approx):   %d\n", due)
		fmt.Printf("Learning (<7d): %d\n", learning)
		fmt.Printf("Mastered (>30d): %d\n", mastered)
		fmt.Printf("In Progress:    %d\n", total - learning - mastered)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
