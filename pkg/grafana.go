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
	dashboardTpl         *mustache.Template
	graphTpl             *mustache.Template
	panelHeight          int
	panelWidthGraph      int
	panelWidthSinglestat int
	rowTpl               *mustache.Template
	singlestatTpl        *mustache.Template
}

func (r *Renderer) Render(db Dashboard, panels []Panel) string {
	panelsRendered := []string{}
	posX := 0
	posY := 0
	for _, p := range panels {
		switch p.Type() {
		case PanelTypeRow:
			row := p.(Row)
			row.PosX = 0
			if posY != 0 {
				posY = posY + r.panelHeight
			}

			row.PosY = posY
			panelsRendered = append(panelsRendered, r.rowTpl.Render(row))
			posX = 0
			// A row always has a height of 1
			posY = posY + 1
		case PanelTypeGraph:
			graph := p.(Graph)
			if (posX + r.panelWidthGraph) > 24 {
				posX = 0
				posY = posY + r.panelHeight
			}

			graph.Height = r.panelHeight
			graph.PosX = posX
			graph.PosY = posY
			graph.Width = r.panelWidthGraph
			panelsRendered = append(panelsRendered, r.graphTpl.Render(graph))
			posX = posX + r.panelWidthGraph
		case PanelTypeSinglestat:
			singlestat := p.(Singlestat)
			if (posX + r.panelWidthSinglestat) > 24 {
				posX = 0
				posY = posY + r.panelHeight
			}

			singlestat.Height = r.panelHeight
			singlestat.PosX = posX
			singlestat.PosY = posY
			singlestat.Width = r.panelWidthSinglestat
			panelsRendered = append(panelsRendered, r.singlestatTpl.Render(singlestat))
			posX = posX + r.panelWidthSinglestat
		}
	}

	db.Panels = strings.Join(panelsRendered, ",")
	return r.dashboardTpl.Render(db)
}

var (
	PanelTypeGraph      = "graph"
	PanelTypeRow        = "row"
	PanelTypeSinglestat = "singlestat"
)

type Dashboard struct {
	Panels    string
	Title     string
	Variables []Variable
}

type Variable struct {
	Datasource string
	HasMore    bool
	Query      string
	Name       string
}

func labelsToVariables(datasource string, labels []string, query string) []Variable {
	variables := []Variable{}
	for i, l := range labels {
		v := Variable{Datasource: datasource, Name: l}
		v.HasMore = i+1 < len(labels)
		v.Query = fmt.Sprintf("label_values(%s, %s)", query, l)
		variables = append(variables, v)
	}

	return variables
}

type Row struct {
	ID    int
	PosX  int
	PosY  int
	Title string
}

func (r Row) Type() string {
	return PanelTypeRow
}

type Graph struct {
	Datasource     string
	Description    string
	Format         string
	HasLegend      bool
	HasThreshold   bool
	Height         int
	ID             int
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
	ID                 int
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
	defaultFormat = "short"
)
