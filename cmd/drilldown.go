package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	v1 "github.com/wndhydrnt/autoboard/pkg"
)

var (
	drilldownCounterChangeFunc string
	drilldownGroupLevel        int
	drilldownSelectors         []string
	drilldownPrefix            string
	drilldownTimeRange         string
)

// drilldownCmd represents the drilldown command
var drilldownCmd = &cobra.Command{
	Args:  cobra.MinimumNArgs(2),
	Use:   "drilldown NAME ENDPOINT",
	Short: "Create a dashbaord that displays all metrics exposed by a service",
	Long: `Create a dashbaord that displays all metrics exposed by a service

Arguments:
NAME: The name of the dashboard in Grafana.

ENDPOINT: The endpoint at which a service exposes its Prometheus metrics, e.g. "http://localhost:9090/metrics".

Flags:
--counter-func: autoboard converts counters into panels that display the change of the metric. This flag allows changing
  which PromQL function to use.

--counter-range: autoboard converts counters into panels that display the change of the metric. This flag allows
  changing which PromQL duration to use, e.g. "1m" or "10m".

--filter: Often an endpoint exposes metrics of different sub-systems, e.g. "go_*" and "prometheus_*". Setting --filter
  to "prometheus_" will only create panels for metrics that start with "prometheus_".

--group-level: autoboard can group a subset of metrics into rows. It does this by splitting each name of a metric by the
  "_" separator. The left-most parts of the name will then be grouped according to the value of group-level.
	Example: go_memstats_alloc_bytes will be put under the row "go_memstats" if group-level is set to 2.
	Setting the value to 0 (the default) disables grouping.

--selector: Selectors are added to the dashbaord as variables. They allow switching between different instances of
  services. This flag can be set multiple times to set multiple selectors.

`,
	Run: func(cmd *cobra.Command, args []string) {
		d := v1.NewDrilldown()
		err := d.Run(cfg, drilldownCounterChangeFunc, args[1], drilldownGroupLevel, drilldownSelectors, args[0], drilldownPrefix, drilldownTimeRange)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	drilldownCmd.Flags().StringVar(&drilldownCounterChangeFunc, "counter-func", "rate", "PromQL function to use in panels that display the change of a counter")
	drilldownCmd.Flags().StringVar(&drilldownTimeRange, "counter-range", "5m", "PromQL range duration to use in panels that display the change of a counter")
	drilldownCmd.Flags().StringVar(&drilldownPrefix, "filter", "", "Filter metrics for which to create panels by their prefix")
	drilldownCmd.Flags().IntVar(&drilldownGroupLevel, "group-level", 0, "Group related metrics in rows")
	drilldownCmd.Flags().StringArrayVar(&drilldownSelectors, "selector", []string{"instance"}, "Add dropdowns to the dashbaord.")
	rootCmd.AddCommand(drilldownCmd)
}
