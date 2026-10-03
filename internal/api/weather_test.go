package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adhocteam/recreation-mcp-server/internal/cache"
	"github.com/adhocteam/recreation-mcp-server/pkg/util"
)

func TestWeatherClient_GetWeatherForecast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forecast" {
			t.Errorf("Expected path /forecast, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(loadTestData(t, "openweather_forecast.json")); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	logger := util.NewLogger("error", "json")
	client := NewWeatherClient(server.URL, "test-api-key",
		util.NewHTTPClient(time.Second, 0, logger),
		cache.NewCache(time.Minute, 10, false), logger)
	forecasts, err := client.GetWeatherForecast(context.Background(), 37.738429, -119.564411, "imperial")
	if err != nil {
		t.Fatal(err)
	}
	wantDates := []string{"2023-10-14T13:00:00Z", "2023-10-14T16:00:00Z", "2023-10-14T19:00:00Z"}
	if len(forecasts) != len(wantDates) {
		t.Fatalf("Expected %d periods, got %d", len(wantDates), len(forecasts))
	}
	for i, forecast := range forecasts {
		encoded, err := json.Marshal(forecast)
		if err != nil {
			t.Fatal(err)
		}
		var output struct {
			Date string `json:"dt"`
		}
		if err := json.Unmarshal(encoded, &output); err != nil {
			t.Fatal(err)
		}
		if output.Date != wantDates[i] {
			t.Errorf("Period %d: expected dt %s, got %s", i, wantDates[i], output.Date)
		}
	}
	if forecasts[0].Temperature.Day != 62.4 || forecasts[2].Weather[0].Description != "light rain" || forecasts[2].Rain != 0.5 {
		t.Error("Forecast values differ from the response")
	}
}
