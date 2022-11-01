package reporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

const (
	consoleDomain = "console.imageflux.jp"
)

type ImageFluxClient struct {
	urlBase       string
	client        *http.Client
	email         string
	password      string
	authenticated bool
}

func NewImageFluxClient(email, password string) (*ImageFluxClient, error) {
	ifc := ImageFluxClient{
		urlBase:       fmt.Sprintf("https://%s", consoleDomain),
		email:         email,
		password:      password,
		authenticated: false,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Debug("failed to init cookie jar")
		return nil, err
	}

	ifc.client = &http.Client{
		Jar: jar,
	}

	log.Debug("succeeded to new ImageFluxClient")

	return &ifc, nil
}

func (ifc ImageFluxClient) BuildURL(path string, query url.Values) (*url.URL, error) {
	uri, err := url.Parse(fmt.Sprintf("%s/%s", ifc.urlBase, path))
	if err != nil {
		log.Debug("failed to parse url", rz.String("path", path))
		return nil, err
	}

	uri.RawQuery = query.Encode()

	return uri, nil
}

func (ifc *ImageFluxClient) Authenticate(ctx context.Context) error {
	if ifc.authenticated {
		log.Debug("already authenticated")
		return nil
	}

	uri, err := ifc.BuildURL("auth/login", url.Values{})
	if err != nil {
		return nil
	}

	form := url.Values{}
	form.Set("email_address", ifc.email)
	form.Set("password", ifc.password)

	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), strings.NewReader(form.Encode()))
	if err != nil {
		log.Debug("failed to creat request", rz.String("url", uri.String()), rz.String("form", form.Encode()))
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	res, err := ifc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Debug("failed to read response body")
			return err
		}

		log.Debug("failed to authenticate", rz.Int("code", res.StatusCode), rz.Bytes("body", body))
		return err
	}

	log.Debug("authentication succeeded")
	ifc.authenticated = true

	return nil
}

func (ifc *ImageFluxClient) Do(req *http.Request) (*http.Response, error) {
	res, err := ifc.client.Do(req)
	if err != nil {
		log.Debug(
			"failed to request",
			rz.String("method", req.Method),
			rz.String("url", req.URL.String()),
		)
		return nil, err
	}

	return res, nil
}

type Statistics struct {
	Ok         bool   `json:"ok"`
	Error      string `json:"error"`
	Statistics struct {
		Summary struct {
			TotalCount   StatisticsSummary `json:"totalCount"`
			CachedCount  StatisticsSummary `json:"cachedCount"`
			FailureCount StatisticsSummary `json:"failureCount"`
			InboundBits  StatisticsSummary `json:"inboundBits"`
			OutboundBits StatisticsSummary `json:"outboundBits"`
			HitRatio     StatisticsSummary `json:"hitRatio"`
		} `json:"summary"`
		Reports           []StatisticsReport           `json:"reports"`
		CumulativeReports []StatisticsCumulativeReport `json:"cumulativeReports"`
	} `json:"statistics"`
}

type StatisticsSummary struct {
	Cur float64 `json:"cur"`
	Min float64 `json:"min"`
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
}

type StatisticsReport struct {
	Time         time.Time `json:"time"`
	TotalCount   float64   `json:"totalCount"`
	CachedCount  float64   `json:"cachedCount"`
	FailureCount float64   `json:"failureCount"`
	InboundBits  float64   `json:"inboundBits"`
	OutboundBits float64   `json:"outboundBits"`
}

type StatisticsCumulativeReport struct {
	Time                 time.Time `json:"time"`
	CachedOutboundBytes  int64     `json:"cachedOutboundBytes"`
	MissedOutboundBytes  int64     `json:"missedOutboundBytes"`
	FailureOutboundBytes int64     `json:"failureOutboundBytes"`
}
