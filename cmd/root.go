package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wndhydrnt/autoboard/pkg/config"
)

var cfgFile string
var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "autoboard",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

	addFlagString(rootCmd, "grafana.address", "http://localhost:3000", "Address of Grafana")
	addFlagString(rootCmd, "grafana.datasource", "", "Datasource to set in queries in Grafana")
	addFlagInt(rootCmd, "grafana.panels.height", 5, "Height of a panel on a dashboard")
	addFlagInt(rootCmd, "grafana.panels.graph.width", 12, "Width of a Graph panel on a dashboard")
	addFlagInt(rootCmd, "grafana.panels.singlestat.width", 6, "Width of a Singlestat panel on a dashboard")
	addFlagString(rootCmd, "log.level", "error", "Log level")
	addFlagString(rootCmd, "templates.dashboard", "", "Path to the template used to render a dashboard")
	addFlagString(rootCmd, "templates.graph", "", "Path to the template used to render a graph")
	addFlagString(rootCmd, "templates.row", "", "Path to the template used to render a row")
	addFlagString(rootCmd, "templates.singlestat", "", "Path to the template used to render a singlestat")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.Parse(cfgFile)
	if err != nil {
		fmt.Printf("Error parsing config: %s", err)
		os.Exit(1)
	}
}

func addFlagInt(cmd *cobra.Command, name string, value int, usage string) {
	cmd.PersistentFlags().Int(name, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}

func addFlagString(cmd *cobra.Command, name string, value string, usage string) {
	cmd.PersistentFlags().String(name, value, usage)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
}
