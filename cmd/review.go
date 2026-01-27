package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/LavenderBridge/spaced-repetition/internal/algorithm"
	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/LavenderBridge/spaced-repetition/internal/models"
	"github.com/spf13/cobra"
	"os/exec"
	"runtime"
)

var reviewOpen bool

var reviewCmd = &cobra.Command{
	Use:   "review [optional problem name]",
	Short: "Start a review session",
	Long: `Start a review session. 
If a problem name is provided, review that specific problem.
If no name provided, review all problems due today.`,
	Run: func(cmd *cobra.Command, args []string) {
		store, err := db.NewStore()
		if err != nil {
			fmt.Println("❌ Database error:", err)
			return
		}
		defer store.Close()

		var problems []models.Problem

		if len(args) > 0 {
			// Review specific problem by exact name match
			name := strings.Join(args, " ")
			p, err := store.GetProblem(name)
			if err != nil {
				fmt.Println("❌ Problem not found or error:", err)
				return
			}
			problems = append(problems, *p)
		} else {
			// Review due problems
			problems, err = store.ListProblems(true) // dueOnly = true
			if err != nil {
				fmt.Println("❌ Error fetching due problems:", err)
				return
			}
			if len(problems) == 0 {
				fmt.Println("✅ No problems due for review today!")
				return
			}
		}

		reader := bufio.NewReader(os.Stdin)

		for i, p := range problems {
			fmt.Println("\n========================================")
			fmt.Printf("Reviewing [%d/%d]: %s\n", i+1, len(problems), p.Name)
			if p.URL != "" {
				fmt.Printf("URL: %s\n", p.URL)
			}
			if p.Notes != "" {
				fmt.Printf("Notes: %s\n", p.Notes)
			}
			fmt.Println("========================================")
			
			if reviewOpen && p.URL != "" {
				fmt.Println("🌐 Opening URL in browser...")
				openBrowser(p.URL)
			}

			fmt.Println("Press Enter to show quality prompt...")
			reader.ReadString('\n')

			fmt.Print("Rate recall quality (0: Blackout -> 5: Perfect): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			
			quality, err := strconv.Atoi(input)
			if err != nil || quality < 0 || quality > 5 {
				fmt.Println("⚠️ Invalid input, skipping update for this problem.")
				continue
			}

			// Update
			updated := algorithm.CalculateReview(p, quality)
			if err := store.UpdateProblem(updated); err != nil {
				fmt.Printf("❌ Error updating problem: %v\n", err)
			} else {
				fmt.Printf("✅ Updated! Next review in %d days.\n", updated.Interval)
			}
		}

		fmt.Println("\n🎉 Review session complete!")
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
	reviewCmd.Flags().BoolVarP(&reviewOpen, "open", "o", false, "Open problem URL in browser")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("❌ Failed to open browser: %v\n", err)
	}
}
