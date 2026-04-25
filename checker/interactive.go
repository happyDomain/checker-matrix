//go:build standalone

package checker

import (
	"errors"
	"net/http"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// RenderForm implements server.Interactive.
func (p *matrixProvider) RenderForm() []sdk.CheckerOptionField {
	return []sdk.CheckerOptionField{
		{
			Id:          "serviceDomain",
			Type:        "string",
			Label:       "Matrix domain",
			Placeholder: "matrix.org",
			Required:    true,
		},
		{
			Id:          "federationTesterServer",
			Type:        "string",
			Label:       "Federation Tester Server",
			Placeholder: "https://federationtester.matrix.org/api/report?server_name=%s",
			Default:     "https://federationtester.matrix.org/api/report?server_name=%s",
			Description: "URL template of the federation tester API; %s is replaced by the domain.",
		},
	}
}

// ParseForm implements server.Interactive.
func (p *matrixProvider) ParseForm(r *http.Request) (sdk.CheckerOptions, error) {
	domain := strings.TrimSpace(r.FormValue("serviceDomain"))
	if domain == "" {
		return nil, errors.New("matrix domain is required")
	}

	opts := sdk.CheckerOptions{
		"serviceDomain": domain,
	}

	if tester := strings.TrimSpace(r.FormValue("federationTesterServer")); tester != "" {
		opts["federationTesterServer"] = tester
	}

	return opts, nil
}
