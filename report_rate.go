package reporter

import (
	"context"
	"encoding/json"
	"fmt"
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
	query := url.Values{}
	query.Set("id", fmt.Sprintf("%d", origin.Id))
	query.Set("gteq", month.StartDate())
	query.Set("lt", month.EndDate())

	uri, err := r.client.BuildURL("statistics/daily", query)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return 0, err
	}

	log.Debug("waiting for response", rz.String("url", uri.String()))
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

	if len(stats.CUMReports) < 1 {
		return int64(0), nil
	}

	latestReport := stats.CUMReports[len(stats.CUMReports)-1]
	log.Debug(
		"report",
		rz.Int("id", origin.Id),
		rz.String("time", latestReport.Time),
		rz.Int64("cached_outbound_bytes", latestReport.CachedOutboundBytes),
		rz.Int64("failure_outbound_bytes", latestReport.FailureOutboundBytes),
		rz.Int64("missed_outbound_bytes", latestReport.MissedOutboundBytes),
	)

	return latestReport.CachedOutboundBytes + latestReport.FailureOutboundBytes + latestReport.MissedOutboundBytes, nil
}
