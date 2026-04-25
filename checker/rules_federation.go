package checker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// federationOKRule reflects the overall FederationOK flag reported by the
// Matrix Federation Tester. Other rules isolate specific concerns; this
// rule is the global verdict so callers get a single-line answer to
// "does this homeserver federate?".
type federationOKRule struct{}

func (r *federationOKRule) Name() string { return "matrix.federation_ok" }
func (r *federationOKRule) Description() string {
	return "Reports the overall federation status returned by the Matrix Federation Tester."
}

func (r *federationOKRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, opts sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	domain, _ := opts["serviceDomain"].(string)
	domain = strings.TrimSuffix(domain, ".")

	if data.FederationOK {
		version := strings.TrimSpace(data.Version.Name + " " + data.Version.Version)
		st := passState("matrix.federation_ok.ok", "Matrix federation is working.")
		if version != "" {
			st.Message = fmt.Sprintf("Matrix federation is working (running %s).", version)
			st.Meta = map[string]any{"version": version}
		}
		return []sdk.CheckState{st}
	}

	var statusLine string
	switch {
	case data.DNSResult.SRVError != nil && data.WellKnownResult.Result != "":
		statusLine = fmt.Sprintf("%s; %s", data.DNSResult.SRVError.Message, data.WellKnownResult.Result)
	case len(data.ConnectionErrors) > 0:
		srvs := make([]string, 0, len(data.ConnectionErrors))
		for srv := range data.ConnectionErrors {
			srvs = append(srvs, srv)
		}
		sort.Strings(srvs)
		var msg strings.Builder
		for _, srv := range srvs {
			if msg.Len() > 0 {
				msg.WriteString("; ")
			}
			msg.WriteString(srv)
			msg.WriteString(": ")
			msg.WriteString(data.ConnectionErrors[srv].Message)
		}
		statusLine = fmt.Sprintf("Connection errors: %s", msg.String())
	default:
		statusLine = fmt.Sprintf("Federation broken. Check https://federationtester.matrix.org/#%s", domain)
	}

	return []sdk.CheckState{critState("matrix.federation_ok.fail", statusLine)}
}
