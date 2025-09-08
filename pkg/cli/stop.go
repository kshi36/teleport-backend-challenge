package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop a running job by ID",
	Long:    "Stop the execution of a running job by providing its job ID.",
	Example: "jobctl stop j-12345",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		message, err := client.StopJob(user, jobID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error stopping job: %v", err)
			return
		}
		fmt.Fprint(cmd.OutOrStdout(), message)
	},
}
