package cli

import (
	"fmt"
	"teleport-jobworker/pkg/jobserver"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop a running job by ID",
	Long:    "Stop the execution of a running job by providing its job ID.",
	Example: "jobctl stop j-12345",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprint(cmd.ErrOrStderr(), errIncorrectArgs)
			return
		}

		jobID := args[0]

		client, err := jobserver.NewClient()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		response, err := client.StopJob(user, jobID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		if response.Error != nil {
			fmt.Fprintf(cmd.OutOrStdout(), messageJobError, *response.Error)
			return
		}

		fmt.Fprintf(cmd.OutOrStdout(), messageJobStopped, response.ID)
	},
}
