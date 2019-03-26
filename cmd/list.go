package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/user"
	"cronitor/lib"
	"github.com/olekukonko/tablewriter"
)

var listCmd = &cobra.Command{
	Use:   "list <optional path>",
	Short: "List all cron jobs found on the server",
	Long: `
Cronitor discover will parse your crontab and create or update monitors using the Cronitor API.

Note: You must supply your Cronitor API key. This can be passed as a flag, environment variable, or saved in your Cronitor configuration file. See 'help configure' for more details.

Example:
  $ cronitor test
      > List cron jobs in your user crontab and system directory
      > Optionally, execute a job and view its output

  $ cronitor test /path/to/crontab
      > Instead of the user crontab, list the jobs in a provided a crontab file (or directory of crontabs)

	`,
	Args: func(cmd *cobra.Command, args []string) error {

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		var username string
		if u, err := user.Current(); err == nil {
			username = u.Username
		}

		crontabs := []*lib.Crontab{}
		commands := []string{}

		if len(args) > 0 {
			// A supplied argument can be a specific file or a directory
			if isPathToDirectory(args[0]) {
				crontabs = lib.ReadCrontabsInDirectory(username, lib.DROP_IN_DIRECTORY, crontabs)
			} else {
				crontabs = lib.ReadCrontabFromFile(username, "", crontabs)
			}
		} else {
			// Without a supplied argument look at the user crontab and the system drop-in directory
			crontabs = lib.ReadCrontabFromFile(username, "", crontabs)
			crontabs = lib.ReadCrontabsInDirectory(username, lib.DROP_IN_DIRECTORY, crontabs)
		}

		if len(crontabs) == 0 {
			printWarningText("No crontab files found")
			return
		}

		fmt.Println()
		for _, crontab := range crontabs {
			if len(crontab.Lines) == 0 {
				continue
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Schedule", "Command", "Monitoring"})
			table.SetAutoWrapText(false)
			table.SetHeaderAlignment(3)

			for _, line := range crontab.Lines {
				if len(line.CommandToRun) == 0 {
					continue
				}

				monitoring := "None"
				if len(line.Code) > 0 || line.HasLegacyIntegration() {
					monitoring = "✔ Ok"
				}

				table.Append([]string{line.CronExpression, line.CommandToRun, monitoring})
				commands = append(commands, line.CommandToRun)
			}

			printSuccessText(fmt.Sprintf("► Reading %s", crontab.DisplayName()))
			table.Render()
			fmt.Println()
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}