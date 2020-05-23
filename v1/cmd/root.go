package cmd

import (
	"flag"
	"fmt"
	"os"

	v1 "github.com/wndhydrnt/autoboard/v1"
	"github.com/wndhydrnt/autoboard/v1/config"
)

var (
	configFilePath = flag.String("config", "", "Path to the configuration file")
)

func Execute() {
	flag.Parse()
	cfg, err := config.Parse(*configFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration file: %s\n", err)
		os.Exit(1)
	}

	err = v1.Run(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
