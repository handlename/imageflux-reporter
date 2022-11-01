package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

type RateReporter struct {
	client  *ImageFluxClient
	origins []Origin
}

func NewRateReporter(ctx context.Context, client *ImageFluxClient, origins []Origin) (*RateReporter, error) {
	if err := client.Authenticate(ctx); err != nil {
		log.Debug("failed to authenticate ImageFluxClient")
		return nil, err
	}

	r := &RateReporter{
		client:  client,
		origins: origins,
	}

	return r, nil
}

type StatisticsPayload struct {
	OriginId int    `json:"originId,omitempty"`
	Interval int    `json:"interval,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
}

type RateReport struct {
	Project string
	Volume  int64
	Rate    float64
}

type RateReports struct {
	Reports map[string]RateReport
}

func NewRateReports() *RateReports {
	return &RateReports{
		Reports: map[string]RateReport{},
	}
}

func (r *RateReports) Add(project string, volume int64) {
	report, exists := r.Reports[project]

	if !exists {
		report = RateReport{
			Project: project,
		}
	}

	report.Volume += volume
	r.Reports[project] = report

	return
}

func (r *RateReports) CalcRate() {
	var total int64 = 0

	for _, report := range r.Reports {
		total += report.Volume
	}

	log.Debug("total volume calculated", rz.Int64("volume", total))

	if total == 0 {
		return
	}

	for project, report := range r.Reports {
		report.Rate = float64(report.Volume) / float64(total)
		r.Reports[project] = report
	}

	return
}

func (r *RateReporter) Run(ctx context.Context, month Month) (*RateReports, error) {
	reports := NewRateReports()

	for _, origin := range r.origins {
		volume, err := r.getOriginTransfers(origin, month, ctx)
		if err != nil {
			log.Debug("failed to get stats", rz.Int("origin", origin.Id))
			return nil, err
		}

		reports.Add(origin.Project, volume)

		time.Sleep(time.Second)
	}

	reports.CalcRate()

	return reports, nil
}

func (r *RateReporter) getOriginTransfers(origin Origin, month Month, ctx context.Context) (int64, error) {
	payload := StatisticsPayload{
		OriginId: origin.Id,
		Interval: 3,
		From:     month.StartDateTime(),
		To:       month.EndDateTime(),
	}

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	uri, err := r.client.BuildURL(".ui-api/statistics.summarized", url.Values{})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), bytes.NewReader(payloadData))
	if err != nil {
		return 0, err
	}

	log.Debug("waiting for response", rz.String("url", uri.String()), rz.Any("payload", payload))
	res, err := r.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debug("failed to read response body")
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		log.Debug("unexpected status code", rz.Int("code", res.StatusCode), rz.Bytes("body", body))
		return 0, err
	}

	var stats Statistics
	if err := json.Unmarshal(body, &stats); err != nil {
		return 0, err
	}

	if len(stats.Statistics.CumulativeReports) < 1 {
		return int64(0), nil
	}

	latestReport := stats.Statistics.CumulativeReports[len(stats.Statistics.CumulativeReports)-1]
	log.Debug(
		"report",
		rz.Int("id", origin.Id),
		rz.Time("time", latestReport.Time),
		rz.Int64("cached_outbound_bytes", latestReport.CachedOutboundBytes),
		rz.Int64("failure_outbound_bytes", latestReport.FailureOutboundBytes),
		rz.Int64("missed_outbound_bytes", latestReport.MissedOutboundBytes),
	)

	return latestReport.CachedOutboundBytes + latestReport.FailureOutboundBytes + latestReport.MissedOutboundBytes, nil
}
