package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	v1 "github.com/wndhydrnt/autoboard/v1"
)

var (
	drilldownPrefix            string
	drilldownTimeRange         string
	drilldownCounterChangeFunc string
)

// drilldownCmd represents the drilldown command
var drilldownCmd = &cobra.Command{
	Args:  cobra.MinimumNArgs(2),
	Use:   "drilldown NAME ENDPOINT",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		d := v1.NewDrilldown()
		err := d.Run(cfg, drilldownCounterChangeFunc, args[1], args[0], drilldownPrefix, drilldownTimeRange)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	drilldownCmd.Flags().StringVar(&drilldownPrefix, "prefix", "", "")
	drilldownCmd.Flags().StringVar(&drilldownTimeRange, "range", "5m", "")
	drilldownCmd.Flags().StringVar(&drilldownCounterChangeFunc, "counter-change-func", "rate", "")
	rootCmd.AddCommand(drilldownCmd)
}
