package checker

import (
	"context"
	"fmt"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Rule returns a new matrix federation check rule.
func Rule() sdk.CheckRule {
	return &matrixRule{}
}

type matrixRule struct{}

func (r *matrixRule) Name() string {
	return "matrix_federation"
}

func (r *matrixRule) Description() string {
	return "Checks whether Matrix federation is working correctly"
}

func (r *matrixRule) ValidateOptions(opts sdk.CheckerOptions) error {
	return nil
}

func (r *matrixRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, opts sdk.CheckerOptions) []sdk.CheckState {
	var data MatrixFederationData
	if err := obs.Get(ctx, ObservationKeyMatrix, &data); err != nil {
		return []sdk.CheckState{{
			Status:  sdk.StatusError,
			Message: fmt.Sprintf("Failed to get Matrix federation data: %v", err),
			Code:    "matrix_federation_error",
		}}
	}

	domain, _ := opts["serviceDomain"].(string)
	domain = strings.TrimSuffix(domain, ".")

	if data.FederationOK {
		version := strings.TrimSpace(data.Version.Name + " " + data.Version.Version)
		return []sdk.CheckState{{
			Status:  sdk.StatusOK,
			Message: fmt.Sprintf("Running %s", version),
			Code:    "matrix_federation_ok",
			Meta: map[string]any{
				"version": version,
			},
		}}
	}

	var statusLine string

	if data.DNSResult.SRVError != nil && data.WellKnownResult.Result != "" {
		statusLine = fmt.Sprintf("%s OR %s", data.DNSResult.SRVError.Message, data.WellKnownResult.Result)
	} else if len(data.ConnectionErrors) > 0 {
		var msg strings.Builder
		for srv, cerr := range data.ConnectionErrors {
			if msg.Len() > 0 {
				msg.WriteString("; ")
			}
			msg.WriteString(srv)
			msg.WriteString(": ")
			msg.WriteString(cerr.Message)
		}
		statusLine = fmt.Sprintf("Connection errors: %s", msg.String())
	} else if data.WellKnownResult.Server != domain {
		statusLine = fmt.Sprintf("Bad homeserver_name: got %s, expected %s", data.WellKnownResult.Server, domain)
	} else {
		statusLine = fmt.Sprintf("Federation broken. Check https://federationtester.matrix.org/#%s", domain)
	}

	return []sdk.CheckState{{
		Status:  sdk.StatusCrit,
		Message: statusLine,
		Code:    "matrix_federation_fail",
	}}
}
