package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

const (
	defaultTesterURI    = "https://federationtester.matrix.org/api/report?server_name=%s"
	collectHTTPTimeout  = 30 * time.Second
	maxResponseBodySize = 5 << 20 // 5 MiB
)

var collectHTTPClient = &http.Client{Timeout: collectHTTPTimeout}

func (p *matrixProvider) Collect(ctx context.Context, opts sdk.CheckerOptions) (any, error) {
	domain, _ := opts["serviceDomain"].(string)
	if domain == "" {
		return nil, fmt.Errorf("serviceDomain is required")
	}
	domain = strings.TrimSuffix(domain, ".")

	testerURI, _ := opts["federationTesterServer"].(string)
	if testerURI == "" {
		testerURI = defaultTesterURI
	}
	if !strings.Contains(testerURI, "%s") {
		return nil, fmt.Errorf("federationTesterServer must contain a %%s placeholder for the domain")
	}

	reqURL := fmt.Sprintf(testerURI, url.QueryEscape(domain))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build the request: %w", err)
	}

	resp, err := collectHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform the test: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("federation tester returned status %d; check https://federationtester.matrix.org/#%s", resp.StatusCode, domain)
	}

	var data MatrixFederationData
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBodySize)).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode federation tester response: %w", err)
	}

	return &data, nil
}
