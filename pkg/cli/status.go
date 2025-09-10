package cli

import (
	"fmt"
	"strconv"
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
			fmt.Fprint(cmd.ErrOrStderr(), errIncorrectArgs)
			return
		}

		jobID := args[0]

		client, err := jobserver.NewClient()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		response, err := client.GetJobStatus(user, jobID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		if response.Error != nil {
			fmt.Fprintf(cmd.OutOrStdout(), messageJobError, *response.Error)
			return
		}

		exitCode := ""
		if response.ExitCode != nil {
			exitCode = strconv.Itoa(*response.ExitCode)
		}

		fmt.Fprintf(cmd.OutOrStdout(), messageJobStatus, response.ID, response.Status, exitCode)
	},
}
