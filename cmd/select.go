package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"github.com/manifoldco/promptui"
	"os/user"
	"cronitor/lib"
)

var selectCmd = &cobra.Command{
	Use:   "select <optional path>",
	Short: "Select a cron job to run interactively",
	Long: `
Cronitor discover will parse your crontab and create or update monitors using the Cronitor API.

Example:
  $ cronitor select
      > List cron jobs in your user crontab and system directory
      > Optionally, execute a job and view its output

  $ cronitor select /path/to/crontab
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

			for _, line := range crontab.Lines {
				if len(line.CommandToRun) == 0 {
					continue
				}

				commands = append(commands, line.CommandToRun)
			}
		}

		prompt := promptui.Select{
			Label: "Select job to run",
			Items: unique(commands),
			Size: 20,
		}

		if _, result, err := prompt.Run(); err == nil {
			if result != "" {
				printSuccessText("► Running command: " + result)
				fmt.Println()
				exitCode := RunCommand(result, false,false)

				if exitCode == 0 {
					printSuccessText("✔ Success")
				} else {
					printErrorText(fmt.Sprintf("✗ Exit code %d", exitCode))
				}

				fmt.Println()
			} else {
				printSuccessText("✔ Done")
			}
		} else if err == promptui.ErrInterrupt {
			printSuccessText("✔ Done")
			os.Exit(-1)
		} else {
			fmt.Println("Error: " + err.Error() + "\n")
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)
}

func unique(strings []string) []string {
    keys := make(map[string]bool)
    list := []string{}
    for _, entry := range strings {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}