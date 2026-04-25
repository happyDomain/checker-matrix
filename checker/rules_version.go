package checker

import (
	"context"
	"fmt"
	"strings"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// versionRule reports whether the federation tester could fetch the
// homeserver version string. The test probe reaches /_matrix/federation/v1/version,
// so a failure here hints at a federation-path problem even when the rest
// of the federation handshake looks healthy.
type versionRule struct{}

func (r *versionRule) Name() string { return "matrix.version" }
func (r *versionRule) Description() string {
	return "Checks that the homeserver responds to /_matrix/federation/v1/version and reports its name and version."
}

func (r *versionRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, _ sdk.CheckerOptions) []sdk.CheckState {
	data, errSt := loadMatrixData(ctx, obs)
	if errSt != nil {
		return []sdk.CheckState{*errSt}
	}

	if data.Version.Error != "" {
		return []sdk.CheckState{warnState("matrix.version.error", fmt.Sprintf("Homeserver /version probe failed: %s", data.Version.Error))}
	}

	version := strings.TrimSpace(data.Version.Name + " " + data.Version.Version)
	if version == "" {
		return []sdk.CheckState{infoState("matrix.version.unknown", "Homeserver did not return a version string.")}
	}

	st := passState("matrix.version.ok", fmt.Sprintf("Homeserver running %s.", version))
	st.Meta = map[string]any{"version": version}
	return []sdk.CheckState{st}
}
