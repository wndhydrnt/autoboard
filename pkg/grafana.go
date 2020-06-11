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
	defaultFormat        = "short"
	grafanaUpdateMessage = "Updated by autoboard"
)

type grafanaCreateDashboardRequest struct {
	Dashboard *json.RawMessage `json:"dashboard"`
	FolderID  int              `json:"folderId"`
	Message   string           `json:"message"`
	Overwrite bool             `json:"overwrite"`
}

// Grafana encapsulates all interactions with the Grafana API.
type Grafana struct {
	Address         string
	FolderIDDefault int
}

// CreateDashboard creates a new dashboard via the Grafana API.
func (g *Grafana) CreateDashboard(d string) error {
	dashboard := json.RawMessage([]byte(d))
	request := &grafanaCreateDashboardRequest{
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

// Renderer contains all logic to create dashboard.
type Renderer struct {
	dashboardTpl         *mustache.Template
	graphTpl             *mustache.Template
	panelHeight          int
	panelWidthGraph      int
	panelWidthSinglestat int
	rowTpl               *mustache.Template
	singlestatTpl        *mustache.Template
}

// Render takes templates, a Dashboard and a list of Panels and creates a JSON data model of dashboard as required by
// Grafana.
// Panels are placed on the dashboard in the order in which they are defined in the slice.
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

// A Panel is a data container.
type Panel interface {
	// Type helps identifying the underlying type of a Panel.
	Type() string
}

// Dashboard holds data used on the top level of the JSON model.
type Dashboard struct {
	Panels    string
	Title     string
	Variables []Variable
}

// Variable is rendered as a selector by Grafana.
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

// A Row is rendered as a row by Grafana.
// Height and width are not configurable because a row in Grafana always has a height of "1" and a width of "24".
type Row struct {
	ID    int
	PosX  int
	PosY  int
	Title string
}

// Type implements Panel.
func (r Row) Type() string {
	return PanelTypeRow
}

// A Graph is rendered as a graph panel by Grafana.
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

// Type implements Panel.
func (g Graph) Type() string {
	return PanelTypeGraph
}

// A GraphQuery is rendered as one query in a graph.
type GraphQuery struct {
	Code    string
	HasMore bool
	Query   string
	RefID   string
}

// A Singlestat is rendered as a singlestat panel by Grafana.
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

// Type implements Panel.
func (s Singlestat) Type() string {
	return PanelTypeSinglestat
}
