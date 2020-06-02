package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/autoboard/pkg/config"
)

type grafanaSearchResponse struct {
	UID string `json:"uid"`
}

type grafanaGetDashboardResponse struct {
	Dashboard map[string]interface{} `json:"dashboard"`
}

func TestAlert(t *testing.T) {
	cfg, err := config.Parse("../test/config.yml")
	require.NoError(t, err)
	err = RunAlert(cfg, []*regexp.Regexp{regexp.MustCompile(".*")})
	require.NoError(t, err)

	actual := readGrafanaDashboard("TestPanels", t)
	// Reset fields dynamically set by Grafana
	actual.Dashboard["uid"] = "abc"
	actual.Dashboard["version"] = float64(1)

	expectedData, err := ioutil.ReadFile("../test/result_alert.json")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual.Dashboard)
}

func TestDrilldown(t *testing.T) {
	cfg, err := config.Parse("../test/config.yml")
	require.NoError(t, err)
	dd := NewDrilldown()
	err = dd.Run(cfg, "rate", "http://127.0.0.1:12958/metrics", "Drilldown Unit Test", "prometheus_tsdb_", "5m")
	require.NoError(t, err)
	actual := readGrafanaDashboard("Drilldown Unit Test", t)
	// Reset fields dynamically set by Grafana
	actual.Dashboard["uid"] = "abc"
	actual.Dashboard["version"] = float64(1)
	expectedData, err := ioutil.ReadFile("../test/result_drilldown.json")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = json.Unmarshal(expectedData, &expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual.Dashboard)
}

func readGrafanaDashboard(q string, t *testing.T) grafanaGetDashboardResponse {
	searchResp, err := http.Get("http://admin:admin@127.0.0.1:12959/api/search?query=" + url.QueryEscape(q))
	require.NoError(t, err)

	b, err := ioutil.ReadAll(searchResp.Body)
	require.NoError(t, err)

	defer searchResp.Body.Close()
	var search []grafanaSearchResponse
	err = json.Unmarshal(b, &search)
	require.NoError(t, err)

	if len(search) != 1 {
		t.Errorf("grafana search returned %d dashboards instead of 1", len(search))
	}

	dashboardResp, err := http.Get("http://admin:admin@127.0.0.1:12959/api/dashboards/uid/" + search[0].UID)
	require.NoError(t, err)

	dashboardPayload, err := ioutil.ReadAll(dashboardResp.Body)
	require.NoError(t, err)

	defer dashboardResp.Body.Close()
	var result grafanaGetDashboardResponse
	err = json.Unmarshal(dashboardPayload, &result)
	require.NoError(t, err)

	return result
}
