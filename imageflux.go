package reporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

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
	CUMReports []CUMReport `json:"cum_reports,omitempty"`
}

type CUMReport struct {
	Time                 string `json:"time,omitempty"`
	CachedOutboundBytes  int64  `json:"cached_outbound_bytes,omitempty"`
	FailureOutboundBytes int64  `json:"failure_outbound_bytes,omitempty"`
	MissedOutboundBytes  int64  `json:"missed_outbound_bytes,omitempty"`
}
