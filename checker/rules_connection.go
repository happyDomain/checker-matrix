package checker

import (
	"context"
	"fmt"
	"sort"

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
		return []sdk.CheckState{unknownState("matrix.connection_reachable.unknown", "No endpoint was probed by the federation tester.")}
	}

	if len(data.ConnectionErrors) == 0 {
		return []sdk.CheckState{passState("matrix.connection_reachable.ok", fmt.Sprintf("All %d endpoint(s) accepted the connection.", len(data.ConnectionReports)))}
	}

	addrs := make([]string, 0, len(data.ConnectionErrors))
	for addr := range data.ConnectionErrors {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)

	out := make([]sdk.CheckState, 0, len(addrs))
	for _, addr := range addrs {
		st := critState("matrix.connection_reachable.fail", data.ConnectionErrors[addr].Message)
		st.Subject = addr
		out = append(out, st)
	}
	return out
}
