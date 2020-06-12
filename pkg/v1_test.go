package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	cfg.GrafanaPassword = "admin"
	cfg.GrafanaUsername = "admin"
	cfg.Datasource = "test_datasource"
	err = RunAlert(cfg, []*regexp.Regexp{regexp.MustCompile(".*")}, "http://localhost:12958", "ab_")
	require.NoError(t, err)

	actual := readGrafanaDashboard("TestPanels", t)
	cleanVariableData(actual.Dashboard)

	expectedData, err := ioutil.ReadFile("../test/result_alert.json")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = json.Unmarshal(expectedData, &expected)
	cleanVariableData(expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual.Dashboard)
}

func TestDrilldown(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile("../test/metrics.txt")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	defer s.Close()
	cfg, err := config.Parse("../test/config.yml")
	require.NoError(t, err)
	cfg.GrafanaFolder = "Test Folder"
	cfg.GrafanaPassword = "admin"
	cfg.GrafanaUsername = "admin"
	dd := NewDrilldown()
	err = dd.Run(cfg, "rate", s.URL+"/metrics", 1, []string{"instance"}, "Drilldown Unit Test", "", "5m")
	require.NoError(t, err)
	actual := readGrafanaDashboard("Drilldown Unit Test", t)
	cleanVariableData(actual.Dashboard)
	expectedData, err := ioutil.ReadFile("../test/result_drilldown.json")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = json.Unmarshal(expectedData, &expected)
	cleanVariableData(expected)
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

// cleanVariableData removes or resets data that can be different between test runs.
func cleanVariableData(d map[string]interface{}) {
	d["id"] = float64(1)
	d["schemaVersion"] = float64(1)
	d["uid"] = "abc"
	d["version"] = float64(1)

	panels := d["panels"].([]interface{})
	for _, panel := range panels {
		ttt, ok := panel.(map[string]interface{})
		if !ok {
			continue
		}

		delete(ttt, "id")
	}
}
