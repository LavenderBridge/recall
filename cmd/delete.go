package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/LavenderBridge/spaced-repetition/internal/db"
	"github.com/spf13/cobra"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a problem",
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

		if !forceDelete {
			fmt.Printf("⚠️  Are you sure you want to delete problem %d? (y/N): ", id)
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("❌ Cancelled.")
				return
			}
		}

		if err := store.DeleteProblem(id); err != nil {
			fmt.Println("❌ Error deleting problem:", err)
			return
		}

		fmt.Println("✅ Problem deleted.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation")
}
