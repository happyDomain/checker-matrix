package checker

import (
	"context"
	"fmt"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// connectionReachableRule checks that every federation endpoint returned
// by DNS accepted the TLS connection the tester attempted.
type connectionReachableRule struct{}

func (r *connectionReachableRule) Name() string { return "matrix.connection_reachable" }
func (r *connectionReachableRule) Description() string {
	return "Checks that every discovered federation endpoint accepts an inbound connection."
}

func (r *connectionReachableRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, _ sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	if len(data.ConnectionErrors) == 0 && len(data.ConnectionReports) == 0 {
		return []sdk.CheckState{infoState("matrix.connection_reachable.unknown", "No endpoint was probed by the federation tester.")}
	}

	if len(data.ConnectionErrors) == 0 {
		return []sdk.CheckState{passState("matrix.connection_reachable.ok", fmt.Sprintf("All %d endpoint(s) accepted the connection.", len(data.ConnectionReports)))}
	}

	out := make([]sdk.CheckState, 0, len(data.ConnectionErrors))
	for addr, cerr := range data.ConnectionErrors {
		st := critState("matrix.connection_reachable.fail", cerr.Message)
		st.Subject = addr
		out = append(out, st)
	}
	return out
}
