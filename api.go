package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// rangeMap maps a Range to the range name expected by the stats endpoint.
// "today" and "yesterday" are not listed: they are served by the
// summaries endpoint.
var rangeMap = map[string]string{
	"week":  "last_7_days",
	"month": "last_30_days",
	"half":  "last_6_months",
	"year":  "last_year",
	"all":   "all_time",
}

// breakdown holds the time spent per category, editor, language, machine,
// operating system and project. It is embedded into the response types,
// so its fields are decoded from the same JSON object as theirs.
type breakdown struct {
	Categories       []StatUnit `json:"categories"`
	Editors          []StatUnit `json:"editors"`
	Languages        []StatUnit `json:"languages"`
	Machines         []StatUnit `json:"machines"`
	OperatingSystems []StatUnit `json:"operating_systems"`
	Projects         []StatUnit `json:"projects"`
}

// DayUnit is a single day of the summaries response.
type DayUnit struct {
	breakdown
	GrandTotal struct {
		TotalSeconds float64 `json:"total_seconds"`
	} `json:"grand_total"`
}

// SummuriesResponse is the response of the summaries endpoint.
// It holds one DayUnit per day of the requested period.
type SummuriesResponse struct {
	Data []DayUnit `json:"data"`
}

// StatsResponse is the response of the stats endpoint.
type StatsResponse struct {
	Data struct {
		TotalSeconds float64 `json:"total_seconds"`
		breakdown
	} `json:"data"`
}

// findByName returns the unit with the given name.
// It returns an error if units contains no such unit.
func findByName(units []StatUnit, name string) *StatUnit {
	len := len(units)
	for i := 0; i < len; i++ {
		if units[i].Name == name {
			return &units[i]
		}
	}
	return nil
}

// toWakafetchStats converts b into WakafetchStats with the given total
// time in seconds. Top entries are taken from the first element of each
// list, since the API sorts them by time in descending order.
func (b *breakdown) toWakafetchStats(totalSec float64) WakafetchStats {
	var w WakafetchStats

	w.TotalSeconds = totalSec

	AICat := findByName(b.Categories, "ai coding")
	if AICat == nil {
		w.AIPercentage = 0
	} else {
		w.AIPercentage = AICat.Percent
	}

	if len(b.Projects) > 0 {
		w.TopProject = b.Projects[0].Name
	} else {
		w.TopProject = ""
	}

	if len(b.OperatingSystems) > 0 {
		w.TopOS = b.OperatingSystems[0].Name
	} else {
		w.TopOS = ""
	}

	if len(b.Machines) > 0 {
		w.TopMachine = b.Machines[0].Name
	} else {
		w.TopMachine = ""
	}

	if len(b.Editors) > 0 {
		w.TopEditor = b.Editors[0].Name
	} else {
		w.TopEditor = ""
	}

	w.Languages = b.Languages

	return w
}

// toWakafetchStats converts the response into WakafetchStats.
func (s *StatsResponse) toWakafetchStats() WakafetchStats {
	return s.Data.breakdown.toWakafetchStats(s.Data.TotalSeconds)
}

// towakafetchStats converts the first day of the response into
// WakafetchStats. It returns an error if the response contains no days.
func (s *SummuriesResponse) towakafetchStats() (WakafetchStats, error) {
	if len(s.Data) == 0 {
		return WakafetchStats{}, errors.New("summaries response contains no days")
	}
	return s.Data[0].breakdown.toWakafetchStats(s.Data[0].GrandTotal.TotalSeconds), nil
}

// deserialize decodes the JSON body of a stats response into s.
func (s *StatsResponse) deserialize(rawJSON []byte) error {
	err := json.Unmarshal(rawJSON, s)
	if err != nil {
		return fmt.Errorf("deserialize statistic response: %w", err)
	}
	return nil
}

// deserialize decodes the JSON body of a summaries response into s.
func (s *SummuriesResponse) deserialize(rawJSON []byte) error {
	err := json.Unmarshal(rawJSON, s)
	if err != nil {
		return fmt.Errorf("deserialize statistic response: %w", err)
	}
	return nil
}

// sendGetRequest sends a GET request to url, authenticated with token via
// HTTP Basic auth, and returns the response body. timeout is in seconds.
// If the status is not 200 OK, the returned error holds the response body.
func sendGetRequest(url string, token string, timeout int) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	encodedAPIKey := base64.StdEncoding.EncodeToString([]byte(token))

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encodedAPIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	default:
		return nil, fmt.Errorf("%s", string(body))
	}
}

// isSummariesRange reports whether r is a single-day range served by the
// summaries endpoint. Other ranges are served by the stats endpoint.
func isSummariesRange(r Range) bool {
	return r == "today" || r == "yesterday"
}

// generateUrl builds the request URL for config.Range. If isSummaries is
// true, it requests the summaries endpoint for that single day, otherwise
// the stats endpoint. For a non-WakaTime server the path of the
// WakaTime-compatible API is added. It returns an error if the API URL
// is malformed or the range is unknown.
func generateUrl(config *Config, isSummaries bool) (string, error) {
	u, err := url.Parse(config.APIURL)
	if err != nil {
		return "", err
	}

	if !config.IsWakaTime {
		u = u.JoinPath("/compat/wakatime/v1")
	}

	if isSummaries {
		day := time.Now()
		if config.Range == "yesterday" {
			day = day.AddDate(0, 0, -1)
		}
		date := day.Format(time.DateOnly)

		u = u.JoinPath("/users/current/summaries")

		query := url.Values{}
		query.Set("start", date)
		query.Set("end", date)
		u.RawQuery = query.Encode()
	} else {
		apiRange, ok := rangeMap[string(config.Range)]
		if !ok {
			return "", fmt.Errorf("unknown range: %s", config.Range)
		}

		u = u.JoinPath("/users/current/stats", apiRange)
	}

	return u.String(), nil
}

// GetStats fetches the statistics for config.Range from the API
// and converts them into WakafetchStats.
func GetStats(config *Config) (*WakafetchStats, error) {
	isSummaries := isSummariesRange(config.Range)

	requestString, err := generateUrl(config, isSummaries)
	if err != nil {
		return nil, fmt.Errorf("generate url string: %w", err)
	}

	rawJSON, err := sendGetRequest(requestString, config.APIKey, 5)
	if err != nil {
		return nil, fmt.Errorf("getting stats: %w", err)
	}

	var w WakafetchStats

	if isSummaries {
		var wireStats SummuriesResponse
		err = wireStats.deserialize(rawJSON)
		if err != nil {
			return nil, err
		}

		w, err = wireStats.towakafetchStats()
		if err != nil {
			return nil, err
		}
	} else {
		var wireStats StatsResponse
		err = wireStats.deserialize(rawJSON)
		if err != nil {
			return nil, err
		} else {
			w = wireStats.toWakafetchStats()
		}
	}
	return &w, nil
}
