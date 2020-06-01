package v1

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/textparse"
	log "github.com/sirupsen/logrus"
	"github.com/wndhydrnt/autoboard/v1/config"
)

type Metric struct {
	Type      textparse.MetricType
	Help      []byte
	Name      []byte
	LabelKeys []string
}

type Drilldown struct {
	Converters []MetricConverter
}

func NewDrilldown() *Drilldown {
	return &Drilldown{
		Converters: []MetricConverter{
			&HistogramConverter{},
			&GaugeInfoConverter{},
			&GaugeDerivConverter{},
			&GaugeTimestampConverter{},
			&GaugeConverter{},
			&CounterConverter{},
		},
	}
}

func (d *Drilldown) Run(cfg config.Config, counterChangeFunc, endpoint, title, prefix, timeRange string) error {
	log.SetLevel(cfg.LogLevel)

	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Get(endpoint)
	if err != nil {
		return fmt.Errorf("reading Prometheus endpoint %s: %v", endpoint, err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body of Prometheus endpoint %s: %v", endpoint, err)
	}

	metrics := parseMetrics(b, resp.Header.Get("Content-Type"), prefix)
	db := Dashboard{
		Title: title,
	}
	db.Graphs, db.SingleStats = d.convertMetricsToPanels(metrics, Options{counterChangeFunc, cfg.DatasourceDefault, timeRange})
	db.HasSinglestat = len(db.SingleStats) > 0
	r := &Renderer{
		dashboardTpl:  cfg.TemplateDashboard,
		graphTpl:      cfg.TemplateGraph,
		singlestatTpl: cfg.TemplateSinglestat,
	}
	gf := &Grafana{Address: cfg.GrafanaAddress}
	s := r.Render(db)
	err = gf.CreateDashboard(s)
	if err != nil {
		return fmt.Errorf("create drilldown dashboard: %s", err)
	}

	return nil
}

func (d *Drilldown) convertMetricsToPanels(metrics []Metric, options Options) ([]Graph, []Singlestat) {
	graphs := []Graph{}
	singlestats := []Singlestat{}
	for _, m := range metrics {
		c := findConverter(m, d.Converters)
		if c == nil {
			log.Debugf("no converter found for metric %s (%s)", string(m.Name), m.Type)
			continue
		}

		panels := c.Do(m, options)
		for _, p := range panels {
			switch p.Type() {
			case PanelTypeGraph:
				graphs = append(graphs, p.(Graph))
				break
			case PanelTypeSinglestat:
				singlestats = append(singlestats, p.(Singlestat))
				break
			}
		}
	}

	return graphs, singlestats
}

func parseMetrics(b []byte, contentType string, prefix string) []Metric {
	metrics := []Metric{}
	cm := Metric{}
	p := textparse.New(b, contentType)
	for {
		et, err := p.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		switch et {
		case textparse.EntryType:
			cm.Name, cm.Type = p.Type()
			continue
		case textparse.EntryHelp:
			cm.Name, cm.Help = p.Help()
			continue
		default:
		}

		if cm.Name != nil {
			var lset labels.Labels
			p.Metric(&lset)
			for _, l := range lset {
				if l.Name == labels.MetricName {
					continue
				}

				cm.LabelKeys = append(cm.LabelKeys, l.Name)
			}
			metrics = append(metrics, cm)
			cm = Metric{}
		}
	}

	result := []Metric{}
	prefixB := []byte(prefix)
	for _, m := range metrics {
		if bytes.HasPrefix(m.Name, prefixB) {
			result = append(result, m)
		}
	}

	return result
}

func findConverter(m Metric, converters []MetricConverter) MetricConverter {
	for _, c := range converters {
		if c.Can(m) {
			return c
		}
	}

	return nil
}
