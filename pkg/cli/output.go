package cli

import (
	"fmt"
	"teleport-jobworker/pkg/jobserver"

	"github.com/spf13/cobra"
)

var outputCmd = &cobra.Command{
	Use:     "output",
	Short:   "Get output of job by ID",
	Long:    "Get the standard output and standard error streams of a job by providing its job ID.",
	Example: "jobctl output j-12345",
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
		message, err := client.GetJobOutput(user, jobID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error getting job output: %v", err)
			return
		}
		fmt.Fprint(cmd.OutOrStdout(), message)
	},
}
