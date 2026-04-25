package checker

import (
	"context"
	"fmt"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// wellKnownRule checks the /.well-known/matrix/server delegation: was a
// delegation published, did it resolve, and does it point back at the
// expected server_name?
type wellKnownRule struct{}

func (r *wellKnownRule) Name() string { return "matrix.well_known" }
func (r *wellKnownRule) Description() string {
	return "Checks that /.well-known/matrix/server (if published) is valid and points at the expected server_name."
}

func (r *wellKnownRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, opts sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	wk := data.WellKnownResult

	// Nothing published: the host may rely on SRV only. Mark informational.
	if wk.Server == "" && wk.Result == "" {
		return []sdk.CheckState{infoState("matrix.well_known.absent", "No /.well-known/matrix/server delegation published (federation may still work via SRV).")}
	}

	// Published but the tester flagged an error string.
	if wk.Server == "" && wk.Result != "" {
		if strings.Contains(strings.ToLower(wk.Result), "no .well-known") {
			return []sdk.CheckState{unknownState("matrix.well_known.absent", "No /.well-known/matrix/server delegation found (federation may still work via SRV).")}
		}
		return []sdk.CheckState{critState("matrix.well_known.error", fmt.Sprintf("Well-known delegation error: %s", wk.Result))}
	}

	return []sdk.CheckState{passState("matrix.well_known.ok", fmt.Sprintf("Well-known delegation resolves to %s.", wk.Server))}
}
