package v1

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/wndhydrnt/autoboard/v1/config"
)

func RunAlert(cfg config.Config, filters []*regexp.Regexp) error {
	SetPrefix(cfg.SettingsPrefix)
	log.SetLevel(cfg.LogLevel)
	promapi, err := NewPrometheusAPI(cfg.PrometheusAddress)
	if err != nil {
		return fmt.Errorf("init Prometheus API client: %w", err)
	}

	p := &Prometheus{
		DatasourceDefault: cfg.DatasourceDefault,
		Filters:           filters,
		PromAPI:           promapi,
	}
	boards, err := p.ReadAlerts()
	if err != nil {
		return fmt.Errorf("read alerts from Prometheus: %w", err)
	}

	r := &Renderer{
		dashboard:  cfg.TemplateDashboard,
		graph:      cfg.TemplateGraph,
		singlestat: cfg.TemplateSinglestat,
	}
	gf := &Grafana{Address: cfg.GrafanaAddress}
	for _, d := range boards {
		s := r.Render(d)
		err := gf.CreateDashboard(s)
		if err != nil {
			return fmt.Errorf("create board %s: %w", d.Title, err)
		}

		log.Infof("board %s created", d.Title)
	}

	return nil
}
