package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hoisie/mustache"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	DatasourceDefault            string
	GrafanaAddress               string
	LogLevel                     log.Level
	GrafanaPanelsHeight          int
	GrafanaPanelsGraphWidth      int
	GrafanaPanelsSinglestatWidth int
	PrometheusAddress            string
	SettingsPrefix               string
	TemplateDashboard            *mustache.Template
	TemplateGraph                *mustache.Template
	TemplateRow                  *mustache.Template
	TemplateSinglestat           *mustache.Template
}

func Parse(path string) (cfg Config, _ error) {
	viper.SetEnvPrefix("ab")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBufferString(defaultConfig))
	if err != nil {
		return cfg, fmt.Errorf("parse default config: %w", err)
	}

	if path != "" {
		viper.SetConfigFile(path)
		err = viper.MergeInConfig()
		if err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	}

	logLvl, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		return cfg, fmt.Errorf("parse log level: %w", err)
	}

	dashboardTpl, err := readTemplate("templates.dashboard", dashboardTplDefault)
	if err != nil {
		return cfg, fmt.Errorf("read dashboard template: %w", err)
	}

	graphTpl, err := readTemplate("templates.graph", graphTplDefault)
	if err != nil {
		return cfg, fmt.Errorf("read graph template: %w", err)
	}

	rowTpl, err := readTemplate("templates.row", rowTplDefault)
	if err != nil {
		return cfg, fmt.Errorf("read row template: %w", err)
	}

	singlestatTpl, err := readTemplate("templates.singlestat", singlestatTplDefault)
	if err != nil {
		return cfg, fmt.Errorf("read singlestat template: %w", err)
	}

	return Config{
		DatasourceDefault:            viper.GetString("grafana.datasource_default"),
		GrafanaAddress:               viper.GetString("grafana.address"),
		GrafanaPanelsHeight:          viper.GetInt("grafana.panels.height"),
		GrafanaPanelsGraphWidth:      viper.GetInt("grafana.panels.graph.width"),
		GrafanaPanelsSinglestatWidth: viper.GetInt("grafana.panels.singlestat.width"),
		LogLevel:                     logLvl,
		PrometheusAddress:            viper.GetString("prometheus.address"),
		SettingsPrefix:               viper.GetString("prometheus.settings_prefix"),
		TemplateDashboard:            dashboardTpl,
		TemplateGraph:                graphTpl,
		TemplateRow:                  rowTpl,
		TemplateSinglestat:           singlestatTpl,
	}, nil
}

func readTemplate(cfgKey, def string) (*mustache.Template, error) {
	dashboardTplPath := viper.GetString(cfgKey)
	if dashboardTplPath == "" {
		return mustache.ParseString(def)
	}

	return mustache.ParseFile(dashboardTplPath)
}
