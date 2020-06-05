package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hoisie/mustache"
)

const (
	grafanaUpdateMessage = "Updated by autoboard"
)

type GrafanaCreateDashboardRequest struct {
	Dashboard *json.RawMessage `json:"dashboard"`
	FolderID  int              `json:"folderId"`
	Message   string           `json:"message"`
	Overwrite bool             `json:"overwrite"`
}

type Grafana struct {
	Address         string
	FolderIDDefault int
}

func (g *Grafana) CreateDashboard(d string) error {
	dashboard := json.RawMessage([]byte(d))
	request := &GrafanaCreateDashboardRequest{
		Dashboard: &dashboard,
		FolderID:  g.FolderIDDefault,
		Message:   grafanaUpdateMessage,
		Overwrite: true,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := http.Post(g.Address+"/api/dashboards/db", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read error reading response body: %w", err)
		}

		defer resp.Body.Close()
		return fmt.Errorf("grafana returned %d: %s", resp.StatusCode, b)
	}

	return nil
}

type Renderer struct {
	dashboardTpl  *mustache.Template
	graphTpl      *mustache.Template
	singlestatTpl *mustache.Template
}

func (r *Renderer) Render(db Dashboard, panels []Panel) string {
	panelsRendered := []string{}
	graphWidth := int(24 / graphPanelsPerRow)
	singlestatWidth := int(24 / singleStatPanelsPerRow)
	posX := 0
	posY := 0
	for _, p := range panels {
		switch p.Type() {
		case PanelTypeGraph:
			graph := p.(Graph)
			if (posX + graphWidth) > 24 {
				posX = 0
				posY = posY + panelHeight
			}

			graph.PosX = posX
			graph.PosY = posY
			panelsRendered = append(panelsRendered, r.graphTpl.Render(graph))
			posX = posX + graphWidth
		case PanelTypeSinglestat:
			singlestat := p.(Singlestat)
			if (posX + singlestatWidth) > 24 {
				posX = 0
				posY = posY + panelHeight
			}

			singlestat.PosX = posX
			singlestat.PosY = posY
			panelsRendered = append(panelsRendered, r.singlestatTpl.Render(singlestat))
			posX = posX + singlestatWidth
		}
	}

	db.Panels = strings.Join(panelsRendered, ",")
	return r.dashboardTpl.Render(db)
}

var (
	PanelTypeGraph      = "graph"
	PanelTypeSinglestat = "singlestat"
)

type Dashboard struct {
	Panels string
	Title  string
}

type Graph struct {
	Datasource     string
	Description    string
	Format         string
	HasLegend      bool
	HasThreshold   bool
	Height         int
	Legend         string
	Queries        []GraphQuery
	PosX           int
	PosY           int
	ThresholdOP    string
	ThresholdValue string
	Title          string
	Width          int
}

func (g Graph) Type() string {
	return PanelTypeGraph
}

type GraphQuery struct {
	Code    string
	HasMore bool
	Query   string
	RefID   string
}

type Singlestat struct {
	Datasource         string
	Description        string
	Format             string
	Height             int
	Legend             string
	Query              string
	PosX               int
	PosY               int
	ThresholdInvertNo  bool
	ThresholdInvertYes bool
	ThresholdValue     string
	Title              string
	ValueName          string
	Width              int
}

func (s Singlestat) Type() string {
	return PanelTypeSinglestat
}

const (
	defaultFormat          = "short"
	graphPanelsPerRow      = 2
	panelHeight            = 5
	panelWidth             = 6
	singleStatPanelsPerRow = 4
)
