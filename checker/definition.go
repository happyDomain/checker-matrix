package checker

import (
	"time"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Version is the checker version reported in CheckerDefinition.Version.
//
// It defaults to "built-in", which is appropriate when the checker package is
// imported directly (built-in or plugin mode). Standalone binaries (like
// main.go) should override this from their own Version variable at the start
// of main(), which makes it easy for CI to inject a version with a single
// -ldflags "-X main.Version=..." flag instead of targeting the nested
// package path.
var Version = "built-in"

// Definition returns the CheckerDefinition for the matrix federation checker.
func (p *matrixProvider) Definition() *sdk.CheckerDefinition {
	return &sdk.CheckerDefinition{
		ID:      "matrixim",
		Name:    "Matrix Federation Tester",
		Version: Version,
		Availability: sdk.CheckerAvailability{
			ApplyToService:  true,
			LimitToServices: []string{"abstract.MatrixIM"},
		},
		HasHTMLReport:   true,
		ObservationKeys: []sdk.ObservationKey{ObservationKeyMatrix},
		Options: sdk.CheckerOptionsDocumentation{
			RunOpts: []sdk.CheckerOptionDocumentation{
				{
					Id:          "serviceDomain",
					Type:        "string",
					Label:       "Matrix domain",
					Placeholder: "matrix.org",
					Default:     "matrix.org",
					AutoFill:    sdk.AutoFillDomainName,
					Required:    true,
				},
			},
			AdminOpts: []sdk.CheckerOptionDocumentation{
				{
					Id:          "federationTesterServer",
					Type:        "string",
					Label:       "Federation Tester Server",
					Placeholder: "https://federationtester.matrix.org/api/report?server_name=%s",
					Default:     "https://federationtester.matrix.org/api/report?server_name=%s",
				},
			},
		},
		Rules: Rules(),
		Interval: &sdk.CheckIntervalSpec{
			Min:     5 * time.Minute,
			Max:     7 * 24 * time.Hour,
			Default: 24 * time.Hour,
		},
	}
}
