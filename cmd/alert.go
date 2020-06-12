package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	v1 "github.com/wndhydrnt/autoboard/pkg"
)

var (
	alertPrometheusAddress string
	alertSettingPrefix     string
)

// alertCmd represents the alert command
var alertCmd = &cobra.Command{
	Args:  cobra.MinimumNArgs(1),
	Use:   "alert NAME [NAME...]",
	Short: "Generate a dashboard from an Alert Group in Prometheus",
	Long:  `Generate a dashboard from an Alert Group in Prometheus`,
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

		err := v1.RunAlert(cfg, filters, alertPrometheusAddress, alertSettingPrefix)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			os.Exit(1)
		}
	},
}

func init() {
	alertCmd.Flags().StringVar(&alertPrometheusAddress, "prometheus.address", "http://localhost:9090", "Address of Prometheus")
	alertCmd.Flags().StringVar(&alertSettingPrefix, "setting.prefix", "ab_", "Prefix to identify a setting from annotations of an alert")

	rootCmd.AddCommand(alertCmd)
}
