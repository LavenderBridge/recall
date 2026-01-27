package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "Show problems due for review today",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		problems, err := store.ListProblems(true)
		if err != nil {
			fmt.Println("❌ Error listing due problems:", err)
			return
		}

		if len(problems) == 0 {
			fmt.Println("✅ No problems due today! Good job.")
			return
		}

		fmt.Printf("🔥 %d Problems due today:\n\n", len(problems))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tProblem\tDiff\tNext Review\tTags")
		fmt.Fprintln(w, "--\t-------\t----\t-----------\t----")

		for _, p := range problems {
			var tagNames []string
			for _, t := range p.Tags {
				tagNames = append(tagNames, t.Name)
			}
			tagsStr := strings.Join(tagNames, ", ")

			fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\n", 
				p.ID, p.Name, p.Difficulty, p.NextReview.Format("2006-01-02"), tagsStr)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(dueCmd)
}
