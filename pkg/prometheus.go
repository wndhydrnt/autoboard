package v1

import (
	"context"
	"fmt"
	"regexp"

	promapi "github.com/prometheus/client_golang/api"
	pav1 "github.com/prometheus/client_golang/api/prometheus/v1"
	log "github.com/sirupsen/logrus"
)

var (
	settingPrefix = "ab_"
)

func SetPrefix(p string) {
	settingPrefix = p
}

type Prometheus struct {
	DatasourceDefault string
	Filters           []*regexp.Regexp
	PromAPI           pav1.API
}

func (p *Prometheus) ReadAlerts() (boards []Dashboard, err error) {
	result, err := p.PromAPI.Rules(context.Background())
	if err != nil {
		return boards, fmt.Errorf("read alerts from Prometheus server: %w", err)
	}

	boards = []Dashboard{}
	for _, g := range result.Groups {
		log.Debugf("processing alert group '%s'", g.Name)
		if !p.isAllowed(g.Name) {
			log.Debugf("filtered alert group '%s'", g.Name)
			continue
		}

		board := Dashboard{Title: g.Name}
		for _, rule := range g.Rules {
			ar, ok := rule.(pav1.AlertingRule)
			if !ok {
				continue
			}

			datasource := settingString(ar, "datasource", p.DatasourceDefault)
			metrics, err := ConvertAlertToPanel(ar, datasource)
			if err != nil {
				return boards, fmt.Errorf("convert query to metrics: %w", err)
			}

			switch v := metrics.(type) {
			case Graph:
				v.Title = settingString(ar, "title", ar.Name)
				board.Panels = append(board.Panels, v)
			case Singlestat:
				v.Title = settingString(ar, "title", ar.Name)
				board.Panels = append(board.Panels, v)
			}
		}

		boards = append(boards, board)
	}

	return boards, nil
}

func (p *Prometheus) isAllowed(name string) bool {
	if len(p.Filters) == 0 {
		return false
	}

	for _, f := range p.Filters {
		m := f.MatchString(name)
		if m {
			return m
		}
	}

	return false
}

func settingString(r pav1.AlertingRule, key, def string) string {
	search := settingPrefix + key
	for k, v := range r.Annotations {
		if string(k) == search {
			return string(v)
		}
	}

	return def
}

func NewPrometheusAPI(addr string) (pav1.API, error) {
	c, err := promapi.NewClient(promapi.Config{Address: addr})
	if err != nil {
		return nil, fmt.Errorf("create Prometheus API client: %w", err)
	}

	return pav1.NewAPI(c), nil
}
