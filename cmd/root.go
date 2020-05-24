package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wndhydrnt/autoboard/v1/config"
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
