package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

func (p *matrixProvider) Collect(ctx context.Context, opts sdk.CheckerOptions) (any, error) {
	domain, _ := opts["serviceDomain"].(string)
	if domain == "" {
		return nil, fmt.Errorf("serviceDomain is required")
	}
	domain = strings.TrimSuffix(domain, ".")

	testerURI, _ := opts["federationTesterServer"].(string)
	if testerURI == "" {
		testerURI = "https://federationtester.matrix.org/api/report?server_name=%s"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(testerURI, domain), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build the request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform the test: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("federation tester returned status %d; check https://federationtester.matrix.org/#%s", resp.StatusCode, domain)
	}

	var data MatrixFederationData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode federation tester response: %w", err)
	}

	return &data, nil
}
