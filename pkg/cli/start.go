package cli

import (
	"fmt"
	"teleport-jobworker/pkg/jobserver"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new job",
	Long: `Start a new job by specifying the absolute path to a program and optional arguments.
A new job ID will be returned.`,
	Example: `jobctl start /bin/echo "Hello world!"`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprint(cmd.ErrOrStderr(), errIncorrectArgs)
			return
		}

		program := args[0]
		programArgs := []string{}

		if len(args) > 1 {
			programArgs = args[1:]
		}

		client, err := jobserver.NewClient()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		response, err := client.StartJob(user, program, programArgs)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v", err)
			return
		}

		if response.Error != nil {
			fmt.Fprintf(cmd.OutOrStdout(), messageJobError, *response.Error)
			return
		}

		fmt.Fprintf(cmd.OutOrStdout(), messageJobStarted, response.ID)
	},
}
