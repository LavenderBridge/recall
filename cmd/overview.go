package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/spf13/cobra"
)

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show overview of progress and stats",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		stats, err := store.GetReviewStats()
		if err != nil {
			fmt.Println("❌ Error fetching stats:", err)
			return
		}

		fmt.Println("\n📊 Performance Overview")
		fmt.Println("=======================")
		fmt.Printf("Total Reviews:      %d\n", stats.TotalReviews)
		fmt.Printf("Reviews Last 7D:    %d\n", stats.ReviewsLast7Days)
		fmt.Printf("Average Quality:    %.2f\n", stats.AverageQuality)
		
		fmt.Println("\n📈 Problem Distribution by Difficulty (1-5)")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Difficulty\tCount")
		fmt.Fprintln(w, "----------\t-----")
		
		for i := 1; i <= 5; i++ {
			count := stats.CountByDifficulty[i]
			bar := ""
			for j := 0; j < count; j++ {
				bar += "█"
			}
			fmt.Fprintf(w, "%d\t%d\t%s\n", i, count, bar)
		}
		w.Flush()
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(overviewCmd)
}
