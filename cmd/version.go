package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// BuildDate is the date at which the baniry was built.
	BuildDate = ""
	// BuildHash is git commit hash from which the baniry was built.
	BuildHash = ""
	// Version is the version of the binary.
	Version = "master"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s\n", Version, BuildHash, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
