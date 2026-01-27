package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked problems",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		problems, err := store.ListProblems(false)
		if err != nil {
			fmt.Println("❌ Error listing problems:", err)
			return
		}

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
	rootCmd.AddCommand(listCmd)
}
