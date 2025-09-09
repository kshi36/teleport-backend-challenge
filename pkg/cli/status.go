package cli

import (
	"fmt"
	"teleport-jobworker/pkg/jobserver"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Get status of job by ID",
	Long:    "Get the job status and exit code by providing its job ID.",
	Example: "jobctl status j-12345",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error starting job: incorrect number of args")
			return
		}

		jobID := args[0]

		client, err := jobserver.NewClient("")
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error starting job: %v", err)
			return
		}
		message, err := client.GetJobStatus(user, jobID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error getting job status: %v", err)
			return
		}
		fmt.Fprint(cmd.OutOrStdout(), message)
	},
}
