package cli

import (
	"fmt"
	"os"
	"teleport-jobworker/pkg/jobserver"

	"github.com/spf13/cobra"
)

var client *jobserver.Client
var user string

var rootCmd = &cobra.Command{
	Use:   "jobctl",
	Short: "Manage jobs for Linux processes",
	Long: "jobctl allows users to perform job functions " +
		"on Linux processes over HTTPS: start, stop, get status, get output.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// CLI wraps the jobserver.Client to manage communication with HTTPS API server.
		jsClient, err := jobserver.NewClient("")
		if err != nil {
			return err
		}
		client = jsClient

		return nil
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	// optional user flag to signify to server a different user ID
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "user1", "Assign a user ID to client")

	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(outputCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
}
