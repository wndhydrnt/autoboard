package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	v1 "github.com/wndhydrnt/autoboard/v1"
)

// alertCmd represents the alert command
var alertCmd = &cobra.Command{
	Args:  cobra.MinimumNArgs(1),
	Use:   "alert NAME [NAME...]",
	Short: "Generate a dashboard from an Alert Group in Prometheus",
	Run: func(cmd *cobra.Command, args []string) {
		filters := []*regexp.Regexp{}
		for _, a := range args {
			r, err := regexp.Compile(a)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Create regex from %s: %s\n", a, err)
				os.Exit(1)
			}

			filters = append(filters, r)
		}

		err := v1.RunAlert(cfg, filters)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(alertCmd)
}
